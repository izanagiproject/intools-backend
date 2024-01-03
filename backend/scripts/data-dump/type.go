package main

type Material struct {
	No int `json:"id"`
	Plant          string `json:"plant"`
	Area           string `json:"area"`
	Qcode          string `json:"qcode"`
	ElectricalRoom string `json:"electrical_room"`
	Name           string `json:"name"` //motor name
	Specification Specification `json:"specifications"`
	Size Size `json:"size"`
	Maker string `json:"maker"`
	SerialNumber string `json:"serial_number"`
	PIC PIC `json:"pic"`
	RotorBar RotorBar `json:"rotor_bar"`
	StartingCurrent StartingCurrent `json:"starting_current"`
	Frame int `json:"frame"`
	Type string `json:"type"`
	Installed int8 `json:"installed_qty"`
	StandBy int8 `json:"standby_qty"`
	Spare int8 `json:"spare_qty"`
}

type Specification struct {
	Capacity float32 `json:"capacity"`
	Voltage  float32 `json:"voltage"`
	Current  float32 `json:"current"`
	RPM      float32 `json:"rpm"`
} 

type StartingCurrent struct {
	When string `json:"when"`
	Check string `json:"check"`
}

type RotorBar struct {
	CheckStatus string `json:"check_status"`
	CheckDate string `json:"check_date"`
	Reason string `json:"reason"`
	Remark string `json:"remark"`
}

type Size struct {
	ShaftDiameter float32 `json:"shaft_diameter"`
	BaseWidth     float32 `json:"base_width"`
	BaseLength    float32 `json:"base_length"`
	C             float32 `json:"c"`
	E             float32 `json:"e"`
	H             float32 `json:"h"`
}

type PIC struct {
	Team  string `json:"team"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}