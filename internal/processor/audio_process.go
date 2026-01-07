package processor

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mjibson/go-dsp/fft"
	// Assurez-vous d'avoir exécuté 'go mod tidy' pour obtenir le fichier go.sum
)

// AudioData contient les informations audio brutes reçues du client
type AudioData struct {
	SampleRate uint32 // Fréquence d'échantillonnage (exemple : 44100 Hz)
	Channels   uint16 // Nombre de canaux (1 : Mono, 2 : Stéréo)
	BitDepth   uint16 // Profondeur de bits (exemple : 16 bits)
	Format     uint16 // Format audio
	Samples    []byte // Données brutes sous forme de tableau de bytes (Little-Endian)
}

// AudioChunk est l'unité de données normalisée pour le traitement parallèle
type AudioChunk struct {
	ID         int       // Numéro de séquence du fragment audio pour le réassemblage ultérieur
	Samples    []float64 // Tableau de données converti en nombres réels (-1.0 à 1.0)
	SampleRate float64   // Fréquence d'échantillonnage utilisée pour les calculs FFT
}

// Result contient le résultat après traitement d'un fragment
type Result struct {
	ChunkID int       // ID correspondant à l'AudioChunk initial
	Samples []float64 // Tableau de données après filtrage des fréquences
}

/*
StartProcessor initialise les workers exécutés en parallèle.
Paramètres :
- numWorkers (int) : Nombre de goroutines à exécuter simultanément.
- jobs (<-chan AudioChunk) : Canal pour recevoir les données d'entrée (lecture seule).
- results (chan<- Result) : Canal pour envoyer les résultats de sortie (écriture seule).
*/
func StartProcessor(numWorkers int, jobs <-chan AudioChunk, results chan<- Result) {
	for w := 1; w <= numWorkers; w++ {
		// Exploite les directions des canaux pour un traitement sécurisé
		go audioWorker(w, jobs, results)
	}
}

/*
ProcessAudio est la fonction principale qui orchestre l'ensemble du flux de traitement audio.
Paramètres :
- data (AudioData) : Structure contenant les données brutes et métadonnées du client.
*/
func ProcessAudio(data AudioData) []byte {
	// Convertir les bytes bruts en nombres réels float64
	floats := bytesToFloat64(data.Samples, data.BitDepth, data.Channels)

	// Diviser les données en fragments (chunks) pour le traitement parallèle
	chunkSize := 1024
	numChunks := len(floats) / chunkSize
	if len(floats)%chunkSize != 0 {
		numChunks++
	}

	// Initialiser les canaux de communication entre les goroutines
	jobs := make(chan AudioChunk, numChunks)
	results := make(chan Result, numChunks)

	// Démarrer 4 workers en parallèle
	numWorkers := 4
	StartProcessor(numWorkers, jobs, results)

	// Placer les fragments de données dans la file d'attente jobs
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(floats) {
			end = len(floats)
		}
		chunk := AudioChunk{
			ID:         i,
			Samples:    floats[start:end],
			SampleRate: float64(data.SampleRate),
		}
		jobs <- chunk
	}
	close(jobs) // Fermer le canal après avoir envoyé toutes les données

	// Collecter les résultats des workers
	processedFloats := make([]float64, len(floats))
	for i := 0; i < numChunks; i++ {
		result := <-results
		start := result.ChunkID * chunkSize
		copy(processedFloats[start:], result.Samples) // Copier les données traitées dans le tableau commun
	}

	// Convertir à nouveau de nombres réels en tableau de bytes pour renvoyer au client
	return float64ToBytes(processedFloats, data.BitDepth)
}

/*
audioWorker est la logique d'exécution de chaque goroutine.
Paramètres :
- id (int) : Identifiant du worker (utilisé pour le débogage).
- jobs (<-chan AudioChunk) : Canal pour obtenir les données à traiter.
- results (chan<- Result) : Canal pour retourner les résultats.
*/
func audioWorker(id int, jobs <-chan AudioChunk, results chan<- Result) {
	for chunk := range jobs {
		// Convertir float64 en complex128 car FFT nécessite des nombres complexes
		complexSamples := make([]complex128, len(chunk.Samples))
		for i, s := range chunk.Samples {
			complexSamples[i] = complex(s, 0) // complex(real, imag) est la fonction constructeur pour les nombres complexes en Go
		}

		// Appliquer la transformée de Fourier (passage au domaine fréquentiel)
		spectrum := fft.FFT(complexSamples)

		// Filtre passe-bas : Supprimer les hautes fréquences (3/4 du spectre supérieur)
		for i := len(spectrum) / 4; i < len(spectrum); i++ {
			spectrum[i] = complex(0, 0)
		}

		// Transformée de Fourier inverse (retour au domaine temporel)
		processedSamples := fft.IFFT(spectrum)

		// Extraire la partie réelle des nombres complexes pour obtenir l'onde audio réelle
		realSamples := make([]float64, len(processedSamples))
		for i, s := range processedSamples {
			realSamples[i] = real(s)
		}

		// Envoyer le résultat traité
		results <- Result{
			ChunkID: chunk.ID,
			Samples: realSamples,
		}
	}
}

/*
findDominantFrequency trouve la fréquence la plus forte dans un segment audio.
Paramètres :
- spectrum ([]complex128) : Résultat après exécution de FFT.
- sampleRate (float64) : Vitesse d'échantillonnage audio (Hz).
*/
func findDominantFrequency(spectrum []complex128, sampleRate float64) float64 {
	maxMagnitude := 0.0
	maxIndex := 0
	// Parcourir la moitié du spectre de fréquences (en raison de la symétrie de FFT)
	for i := 1; i < len(spectrum)/2; i++ {
		// Calculer l'amplitude en utilisant l'hypoténuse des parties réelle et imaginaire
		magnitude := math.Hypot(real(spectrum[i]), imag(spectrum[i]))
		if magnitude > maxMagnitude {
			maxMagnitude = magnitude
			maxIndex = i
		}
	}
	// Formule de conversion d'index en Hz : index * SampleRate / Nombre_total_d'échantillons
	return float64(maxIndex) * sampleRate / float64(len(spectrum))
}

/*
frequencyToNote convertit une fréquence (Hz) en nom de note musicale.
Paramètres :
- freq (float64) : Fréquence en Hz à identifier.
*/
func frequencyToNote(freq float64) string {
	if freq <= 0 {
		return "Unknown"
	}
	// Formule MIDI : n = 12*log2(f/440) + 69
	n := 12*math.Log2(freq/440.0) + 69
	noteNumber := int(math.Round(n))

	notes := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	note := notes[noteNumber%12]
	octave := noteNumber/12 - 1
	return fmt.Sprintf("%s%d", note, octave)
}

/*
bytesToFloat64 convertit un tableau de bytes en nombres réels normalisés.
Paramètres :
- samples ([]byte) : Données brutes en bytes.
- bitDepth (uint16) : Profondeur de bits (généralement 16).
- channels (uint16) : Nombre de canaux (1 : Mono, 2 : Stéréo).
*/
func bytesToFloat64(samples []byte, bitDepth uint16, channels uint16) []float64 {
	var floats []float64
	bytesPerSample := int(bitDepth / 8) // Exemple : 16 bits = 2 bytes
	numSamples := len(samples) / bytesPerSample

	for i := 0; i < numSamples; i++ {
		offset := i * bytesPerSample
		var sample int16
		if bitDepth == 16 {
			// Lire 2 bytes en Little Endian (comme dans votre exemple 72 f8)
			sample = int16(binary.LittleEndian.Uint16(samples[offset : offset+2]))
		}
		// Diviser par 32768.0 pour normaliser dans l'intervalle [-1.0, 1.0]
		floats = append(floats, float64(sample)/32768.0)
	}

	// Si stéréo, prendre seulement le canal gauche pour simplifier le calcul des notes
	if channels == 2 {
		mono := make([]float64, len(floats)/2)
		for i := 0; i < len(mono); i++ {
			mono[i] = floats[i*2]
		}
		return mono
	}
	return floats
}

/*
float64ToBytes convertit les nombres réels en bytes bruts.
Paramètres :
- samples ([]float64) : Tableau de nombres réels après traitement.
- bitDepth (uint16) : Profondeur de bits souhaitée.
*/
func float64ToBytes(samples []float64, bitDepth uint16) []byte {
	var bytes []byte
	for _, s := range samples {
		// Convertir float64 en entier 16 bits
		sample := int16(s * 32767.0)
		buf := make([]byte, 2)
		// Écrire selon le standard Little Endian
		binary.LittleEndian.PutUint16(buf, uint16(sample))
		bytes = append(bytes, buf...)
	}
	return bytes
}
