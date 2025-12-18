package audio

type wavMetadata struct {
	SampleRate uint32
	Channels   uint16
	Bitdepth   uint16
	Format     uint16
}

type wavPayloard struct {
	Metadata wavMetadata
	Samples  []int16
}
