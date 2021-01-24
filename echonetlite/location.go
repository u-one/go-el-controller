package echonetlite

// LocationCode represents location code
type LocationCode int32

// Location represents location
type Location struct {
	Code   LocationCode
	Number int32
}

// LocationCodes
const (
	Living LocationCode = iota + 1
	Dining
	Kitchen
	Bathroom
	Lavatory
	Washroom
	Corridor
	Room
	Stairs
	Entrance
	Closet
	Garden
	Garage
	Balcony
	Other
)

func (l LocationCode) String() string {
	switch l {
	case Living:
		return "Living"
	case Dining:
		return "Dining"
	case Kitchen:
		return "Kitchen"
	case Bathroom:
		return "Bathroom"
	case Lavatory:
		return "Lavatory"
	case Corridor:
		return "Corridor"
	case Room:
		return "Room"
	case Stairs:
		return "Stairs"
	case Entrance:
		return "Entrance"
	case Closet:
		return "Closet"
	case Garden:
		return "Garden"
	case Garage:
		return "Garage"
	case Balcony:
		return "Balcony"
	case Other:
		return "Other"
	default:
		return "unknown"
	}
}
