package areas

import "fmt"

// Temporary stubbed data

const (
	DidsburyEast = `{
      "name": "Didsbury East",
      "level": "",
      "code": "E05011362",
      "ancestors": [],
      "siblings": [],
      "children": []
    }`
	Manchester = `{
      "name": "Manchester",
      "level": "",
      "code": "E08000003",
      "ancestors": [],
      "siblings": [],
      "children": []
    }`
	NorthWest = `{
      "name": "North West",
      "level": "",
      "code": "E12000002",
      "ancestors": [],
      "siblings": [],
      "children": []
    }`
	England = `{
      "name": "England",
      "level": "",
      "code": "E92000001",
      "ancestors": [],
      "siblings": [],
      "children": []
    }`
)

func StubbedAncestorAPIResponse(code string) string {
	switch code {
	case "E05011362":
		return fmt.Sprintf(`[%s, %s, %s, %s]`, England, NorthWest, Manchester, DidsburyEast)
	case "E08000003":
		return fmt.Sprintf(`[%s, %s, %s]`, England, NorthWest, Manchester)
	case "E12000002":
		return fmt.Sprintf(`[%s, %s]`, England, NorthWest)
	case "E92000001":
		return fmt.Sprintf(`[%s]`, England)
	default:
		return `[]`
	}
}
