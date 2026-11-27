package workflow

type Step struct {
	Sequence string
	Name     string
	Index    int
}

type StepStage string

const (
	StepStageStart  StepStage = "start"
	StepStageFinish StepStage = "finish"
)

type StepEvent struct {
	Step  Step
	Stage StepStage
	Err   error
}
