package trips

import (
	"fmt"
	"testing"
)

// holiday describes the start and end dates for a holiday in string
// format (using 2006-01-02 format)
type holiday struct {
	start string
	end   string
}

// overlap detection checks
func TestTripAdditions(t *testing.T) {

	window := 5
	compoundStayMaxLength := 3

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	type holidayTest struct {
		start string
		end   string
		err   bool
		msg   string
	}
	for i, h := range []holidayTest{
		holidayTest{"2023-01-01", "2023-01-01", false, "simple add"},
		holidayTest{"2023-01-01", "2023-01-07", true, "overlap date"},
		holidayTest{"2023-01-02", "2023-01-04", false, "simple add again"},
		holidayTest{"2023-01-03", "2023-01-05", true, "overlap 2"},
		holidayTest{"2023-01-04", "2023-01-05", true, "overlap 3"},
		holidayTest{"2022-12-31", "2023-01-01", true, "overlap 4"},
		holidayTest{"2022-12-31", "2022-12-30", true, "end before start"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if (err != nil && !h.err) || (err == nil && h.err) {
			t.Errorf("addtrip error for test %d : %v", i, h)
		}
	}
}

// TestTrips checks for a 5 day window over a maximum stay length of 4
// days starting on 1 January
//
//	x....xx...xx..x.....xx.xx
//	1    2    3         4       <- max stay
func TestTrips(t *testing.T) {

	window := 5
	compoundStayMaxLength := 4
	resultsNo := 4

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	for i, h := range []holiday{
		holiday{"2023-01-01", "2023-01-01"},
		holiday{"2023-01-06", "2023-01-07"},
		holiday{"2023-01-11", "2023-01-12"},
		holiday{"2023-01-15", "2023-01-15"},
		holiday{"2023-01-21", "2023-01-22"},
		holiday{"2023-01-24", "2023-01-25"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if err != nil {
			t.Fatalf("error making holiday %d %v %v", i, h, err)
		}
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	fmt.Println(trips)

	breach, windows := trips.LongestTrips(resultsNo)
	fmt.Printf("breach : %t\n", breach)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	if trips.longestStay != 4 {
		t.Errorf("Expected longest stay to be 4, got %d", trips.longestStay)
	}

	if breach != false {
		t.Error("Expected breach to be false, got true")
	}

}

// the same test as above, but with a lower compoundStayMaxLength to
// breach
func TestTripsToBreach(t *testing.T) {

	window := 5
	compoundStayMaxLength := 3
	resultsNo := 4

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	for i, h := range []holiday{
		holiday{"2023-01-01", "2023-01-01"},
		holiday{"2023-01-06", "2023-01-07"},
		holiday{"2023-01-11", "2023-01-12"},
		holiday{"2023-01-15", "2023-01-15"},
		holiday{"2023-01-21", "2023-01-22"},
		holiday{"2023-01-24", "2023-01-25"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if err != nil {
			t.Fatalf("error making holiday %d %v %v", i, h, err)
		}
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	fmt.Println(trips)

	breach, windows := trips.LongestTrips(resultsNo)
	fmt.Printf("breach : %t\n", breach)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	if trips.longestStay != 4 {
		t.Errorf("Expected longest stay to be 4, got %d", trips.longestStay)
	}

	if breach != true {
		t.Error("Expected breach to be true, got false")
	}

}

func TestTripsLonger(t *testing.T) {

	trips, err := NewTrips(40, 35)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	err = trips.AddTrip("2023-02-01", "2023-02-28")
	if err != nil {
		t.Fatalf("1. unexected error making holiday %v", err)
	}

	err = trips.AddTrip("2023-04-02", "2023-05-10")
	if err != nil {
		t.Fatalf("2. unexected error making holiday %v", err)
	}

	err = trips.AddTrip("2022-04-02", "2022-04-01")
	if err == nil {
		t.Fatal("should error with overlapping dates")
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	fmt.Println(trips)

	breach, windows := trips.LongestTrips(5)
	fmt.Printf("breach : %t\n", breach)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	// start 2023-04-02 end 2023-05-10 is 39 days inclusive
	if trips.longestStay != 39 {
		t.Errorf("Expected longest stay to be 39, got %d", trips.longestStay)
	}

	if breach != true {
		t.Error("Expected breach to be true, got false")
	}
}
