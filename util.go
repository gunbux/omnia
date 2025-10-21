package main

// Generic Helper Functions

// Gets launcher box width from the terminal window width
func getBoxWidth(windowWidth int) int {
	// NOTE: -4 to account for padding and border on both sides
	return min(windowWidth-4, MaxBoxWidth)
}
