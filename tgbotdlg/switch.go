package tgbotdlg

type Switch struct {
	name         string
	initialState interface{}
}

func SwitchTo[DialogType dialog[State], State any]() *Switch {
	var d DialogType
	return &Switch{name: d.Name(), initialState: d.newState()}
}

func SwitchWithStateTo[DialogType dialog[State], State any](state State) *Switch {
	var d DialogType
	return &Switch{name: d.Name(), initialState: &state}
}
