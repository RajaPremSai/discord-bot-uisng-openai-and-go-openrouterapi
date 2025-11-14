package dalle

import "fmt"

type imageCommandOptionType uint8

const (
	imageCommandOptionPrompt  imageCommandOptionType = 1
	imageCommandOptionModel   imageCommandOptionType = 2
	imageCommandOptionSize    imageCommandOptionType = 3
	imageCommandOptionNumber  imageCommandOptionType = 4
	imageCommandOptionQuality imageCommandOptionType = 5
	imageCommandOptionStyle   imageCommandOptionType = 6
)

func (t imageCommandOptionType) String() string {
	switch t {
	case imageCommandOptionPrompt:
		return "prompt"
	case imageCommandOptionModel:
		return "model"
	case imageCommandOptionSize:
		return "size"
	case imageCommandOptionNumber:
		return "number"
	case imageCommandOptionQuality:
		return "quality"
	case imageCommandOptionStyle:
		return "style"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}
