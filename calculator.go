package main

// gradeCalculator will start its own input loop
func gradeCalculator(ticket, studentID string) error {
	// this will have a lot of shared code with [showAllGrades]
	if ticket == "" {
		return ErrNotSignedIn
	}

	data, err := getFullDecodedResponse(ticket, studentID)
	if err != nil {
		return err
	}

	quarterStart, quarterEnd := data.Response.Return.Data.GetCurrentQuarter()
	classes, weightIDs := extractInfoFromResponse(data, quarterStart, quarterEnd)

	return nil
}
