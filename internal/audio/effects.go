package audio

import "fmt"

func SoundTrack(id int) string { return fmt.Sprintf("SFX_%02d", id) }

func NewSoundLibrary(dir string) *Library {
	names := make([]string, 28)
	for i := range names {
		names[i] = SoundTrack(i)
	}
	return newLibrary(dir, names)
}
