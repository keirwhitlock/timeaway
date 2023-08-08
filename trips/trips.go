package trips

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// trip is a simple description of a trip with start and end date. The
// trip struct is also used to describe partial trips for windows.trip
type trip struct {
	Start time.Time
	End   time.Time
}

// String returns a string representation of a trip
func (t trip) String() string {
	return fmt.Sprintf(
		"start %s end %s", dayFmt(t.Start), dayFmt(t.End),
	)
}

// days returns the number of inclusive days between the start and end
// dates of a trip
func (t trip) days() int {
	days := 0
	for d := t.Start; !d.After(t.End); d = d.Add(durationDays(1)) {
		days++
	}
	return days
}

// overlap returns a pointer to a partial or full trip if there is an
// overlap with the provided dates, else a nil pointer
func (t trip) overlaps(start, end time.Time) *trip {
	partialTrip := trip{}
	// no overlap
	if t.Start.After(end) || t.End.Before(start) {
		return nil
	}
	// contained
	if t.Start.After(start) && t.End.Before(end) {
		partialTrip.Start = t.Start
		partialTrip.End = t.End
		return &partialTrip
	}
	// partial overlap
	if t.Start.Before(start) || t.Start == start {
		partialTrip.Start = start
	} else {
		partialTrip.Start = t.Start
	}
	if t.End.After(end) || t.End == end {
		partialTrip.End = end
	} else {
		partialTrip.End = t.End
	}
	return &partialTrip
}

// window stores the results of a calculation window
type window struct {
	start     time.Time
	end       time.Time
	tripParts []trip // parts of any overlapping trips
	daysAway  int    // days away for this window
}

// String returns a printable version of a window
func (w window) String() string {
	tpl := `%s : %s (%d)`
	s := fmt.Sprintf(
		tpl, dayFmt(w.start), dayFmt(w.end), w.daysAway,
	)
	for _, t := range w.tripParts {
		s = s + fmt.Sprintf(" %s", t)
	}
	return s
}

// Trips describe a set of trips and other metadata
type Trips struct {
	window      int       // window of days to search over
	maxStay     int       // the maximum length of trips in window
	startFrame  time.Time // date at which to start calculating windows
	endFrame    time.Time // date at which to stop calculating windows
	longestStay int       // the longest compound stay in days
	trips       []trip
	windows     []window
	breach      bool
}

// String returns a simple string representation of trips
func (trips Trips) String() string {
	tpl := `
		window      %d
		maxStay     %d
		startFrame  %s
		endFrame    %s
		longestStay %d
		trips       %d
		windows     %d
		breach      %t`
	tpl = strings.ReplaceAll(tpl, "\t", "")
	return fmt.Sprintf(
		tpl,
		trips.window,
		trips.maxStay,
		dayFmt(trips.startFrame),
		dayFmt(trips.endFrame),
		trips.longestStay,
		len(trips.trips),
		len(trips.windows),
		trips.breach,
	)
}

// NewTrips makes a new Trips struct. The window and maxStay are
// specified in days
func NewTrips(window, maxStay int) (*Trips, error) {
	trips := Trips{}
	trips.breach = false
	if window < 3 {
		return &trips, errors.New("window cannot be less than 3 days")
	}
	if maxStay < 2 {
		return &trips, errors.New("maximum stay cannot be less than 2 days")
	}
	if maxStay > window {
		return &trips, errors.New("maximum stay cannot be greater than the window")
	}
	trips.window = window
	trips.maxStay = maxStay
	return &trips, nil
}

// AddTrip adds a trip to Trips, checking for validity and overlaps
func (trips *Trips) AddTrip(start, end string) error {
	f := func(s string) (time.Time, error) {
		return time.Parse("2006-01-02", s)
	}
	var t trip
	var err error
	t.Start, err = f(start)
	if err != nil {
		return err
	}
	t.End, err = f(end)
	if err != nil {
		return err
	}

	// check validity of this trip
	if t.End.Before(t.Start) {
		return fmt.Errorf("start date %s after %s", t.Start, t.End)
	}
	// check no overlaps
	for _, o := range trips.trips {
		if ok := o.overlaps(t.Start, t.End); ok != nil {
			return fmt.Errorf(
				"trip %s:%s overlaps with %s:%s",
				start, end, dayFmt(o.Start), dayFmt(o.End),
			)
		}
	}

	// set window dates
	x := trip{}
	if trips.startFrame == x.Start || trips.startFrame.After(t.Start) {
		trips.startFrame = t.Start
	}
	if trips.endFrame.Before(t.End) {
		trips.endFrame = t.End
	}

	trips.trips = append(trips.trips, t)
	return nil
}

// Calculate calculates the trip stays for each applicable window
// between the start and end date frames. The window calculator could be
// moved to goroutines to speed up processing, although it seems
// sufficiently fast already.
func (trips *Trips) Calculate() error {
	if len(trips.trips) == 0 {
		return errors.New("no trips were provided to calculate")
	}

	// set suitable frame start and end in which to calculate windows
	windowDuration := durationDays(trips.window - 1) // remove last day
	trips.endFrame = trips.endFrame.Add(-windowDuration)
	if trips.endFrame.Before(trips.startFrame) {
		trips.endFrame = trips.startFrame
	}

	// generate a series of windows starting on each day between
	// trips.startFrame and trips.endFrame and store the results in
	// trips.windows. This loop could be moved to a set of goroutines
	// although peformance for very large windows is still very quick,
	// around 0.005s for a 720 day/180 stay use case.
	for d := trips.startFrame; !d.After(trips.endFrame); d = d.Add(durationDays(1)) {
		w := window{}
		w.start = d
		w.end = d.Add(windowDuration)
		// testStub(d, w)
		for _, t := range trips.trips {
			partialTrip := t.overlaps(w.start, w.end)
			if partialTrip == nil {
				continue
			}
			w.tripParts = append(w.tripParts, *partialTrip)
			w.daysAway += partialTrip.days()
		}
		trips.windows = append(trips.windows, w)
		if w.daysAway > trips.longestStay {
			trips.longestStay = w.daysAway
		}
		if w.daysAway > trips.maxStay {
			trips.breach = true
		}
	}
	return nil
}

// LongestTrips returns a boolean notifying of a breach of the provided
// window and stays parameters together with the analysis windows with
// the longest compound stays, returning at most resultsNo results.
// Normally only the top result is expected to be needed, but note that
// for windows of equal daysAway values, the one dated earliest will
// come first as windows are made in date order.
func (trips *Trips) LongestTrips(resultsNo int) (breach bool, windows []window) {
	breach = trips.breach
	for _, w := range trips.windows {
		if w.daysAway > 0 {
			windows = append(windows, w)
		}
	}
	sort.SliceStable(windows, func(i, j int) bool {
		return windows[i].daysAway > windows[j].daysAway
	})
	if len(windows) >= resultsNo {
		windows = windows[:resultsNo]
	}
	return
}

// durationDays returns a duration for the number of days specified
func durationDays(d int) time.Duration {
	return time.Duration(d) * time.Hour * 24
}

// dayFmt returns the 2006-01-02 representation of a date
func dayFmt(d time.Time) string {
	return d.Format("2006-01-02")
}

// testStub for checking window sizes to be ignored
func testStub(d time.Time, w window) {
	testDay, _ := time.Parse("2006-01-02", "2023-01-11")
	if d.Equal(testDay) {
		fmt.Println("2023-01-11")
		fmt.Printf("day : %s\n", d.Format("2006-01-02"))
		fmt.Printf("window: %s - %s", w.start.Format("2006-01-02"), w.end.Format("2006-01-02"))
		os.Exit(1)
	}
}
