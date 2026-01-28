# ELP-G25 - Audio proccessing client/serveur en Go

Projet Go - groupe 25 (ELP)
Ce projet implémente un client TCP qui lit un ficher `.wav`, envoie le fichier à un serveur TCP. Ce serveur applique un traitement audio puis le renvoie au client qui renvoie un nouveau fichier `.wav`.

## Prérequis
- Go 1.25.5

## Utilisation

### Installation
```sh
git clone https://github.com/Shaa42/ELP-G25.git
```

### Lancement

1. Lancer le serveur dans un terminal :
```sh
go run ./cmd/server
```

2. Puis dans un autre terminal lancer le client :
```sh
go run ./cmd/client
```

Par défaut le client lit `assets/sample-3s.wav`, se connecte à `localhost:42069` puis envoie le fichier. Puis écrit le résultat dans `output.wav`


## Structure du projet
```
.
├── assets              # Dossier des fichiers audio
├── bin
├── cmd
│   ├── client          # Client TCP
│   └── server          # Serveur TCP
├── go.mod
├── go.sum
├── internal
│   ├── audio           # Package gestion des fichiers audio .wav
│   └── processor       # Package traitement audio
├── pkg
├── README.md
└── tests
```

## Points d'amélioration
- Ajouter des options en ligne de commande pour le client (fichier d'entrée, fichier de sortie)
- Traitement des chunks par le serveur en parallèle
- Ajouter des tests
- Ajouter plus de fonctions pour le traitement audio
