package testhelpers

import "github.com/steinfletcher/apitest"

// RecordingFormatter is an apitest.ReportFormatter that records all events
// and can be used to generate a unified report at the end of a test.
type RecordingFormatter struct {
	recorder *apitest.Recorder
}

// NewRecordingFormatter creates a new RecordingFormatter
func NewRecordingFormatter() *RecordingFormatter {
	return &RecordingFormatter{recorder: apitest.NewTestRecorder()}
}

// Format implements the apitest.ReportFormatter interface
func (rf *RecordingFormatter) Format(recorder *apitest.Recorder) {
	// append the events to the existing events
	rf.recorder.Events = append(rf.recorder.Events, recorder.Events...)
}

// SetTitle sets the title of the report
func (rf *RecordingFormatter) SetTitle(title string) {
	rf.recorder.AddTitle(title)
}

// GetRecorder returns the underlying recorder, which can be used to generate
// a unified report at the end of a test.
func (rf *RecordingFormatter) GetRecorder() *apitest.Recorder {
	return rf.recorder
}

// SetOutFileName sets the output file name of the report. This will override
// all internal metadata set on the recorder. The file name will be extended
// with '.html'.
func (rf *RecordingFormatter) SetOutFileName(fileNameWithoutExt string) {
	meta := make(map[string]interface{})
	meta["hash"] = fileNameWithoutExt

	rf.recorder.AddMeta(meta)
}
