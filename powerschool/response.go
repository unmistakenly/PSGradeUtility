package powerschool

import "time"

func (r *DataResponse) GetCurrentQuarter() (start, end time.Time) {
	// this part is like, kinda sucky? i mean not really,
	// we just have to make sure we're currently IN the quarter,
	// and thats it's longer than 2 months but not longer than 3 (year-long classes)
	// sure wouldve been nice if we had definitive data from the api, but im asking for too much!
	now := time.Now().UTC()

	for _, t := range r.Terms {
		start, _ = time.Parse(TimeFormat, t.StartDate)
		end, _ = time.Parse(TimeFormat, t.EndDate)
		if now.Before(start) || now.After(end) {
			continue
		} else if start.AddDate(0, 2, 0).Before(end) && start.AddDate(0, 3, 0).After(end) {
			return
		}
	}

	return // hopefully we wont get to this point
}
