package variants

import "github.com/setnicka/smf-jezero/game"

// Get variant by name
func Get(name string) game.Variant {
	switch name {
	case "coral-reef":
		return NewCoralReef()
	case "hotel":
		return NewHotel()
	case "small-gods":
		return NewSmallGods()
	}
	return nil
}
