package main

type Device struct {
	Alias string
	UUID  int
}

type Watchgroup struct {
	Name     string
	Overseer string
	Devices  []Device
}

type Notification struct {
	ID          int
	Header      string
	Description string
	Status      string
}
