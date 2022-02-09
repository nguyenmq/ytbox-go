package backend

import (
	"testing"
)

var testLinks = []string{
	"https://www.youtube.com/watch?v=SilKjJ0S904",
	"https://www.youtube.com/watch?v=aatr_2MstrI",
	"https://www.youtube.com/watch?v=lMinM-FphYQ",
	"https://www.youtube.com/watch?v=cHkDZ1ekB9U&list=RDcHkDZ1ekB9U&start_radio=1",
	"https://www.youtube.com/watch?v=vjAxLbmy83E",
	"https://www.youtube.com/watch?v=_sV0S8qWSy0",
	"https://youtu.be/ed0CcFcBBMI",
	"https://youtu.be/bL_NcoCJgzo",
	"https://youtu.be/A1oxh8Z-2ko",
	"https://m.youtube.com/watch?v=VQa9Q5_Dcck",
	"https://google.com",
}

var expectedIds = []string{
	"SilKjJ0S904",
	"aatr_2MstrI",
	"lMinM-FphYQ",
	"cHkDZ1ekB9U",
	"vjAxLbmy83E",
	"_sV0S8qWSy0",
	"ed0CcFcBBMI",
	"bL_NcoCJgzo",
	"A1oxh8Z-2ko",
	"VQa9Q5_Dcck",
	"",
}

func TestExtractVideoId_when_success(t *testing.T) {
	for index, link := range testLinks {
		result := extractVideoId(link)

		if result != expectedIds[index] {
			t.Errorf("Expected id %s should match actual id %s\n", expectedIds[index], result)
		}
	}
}
