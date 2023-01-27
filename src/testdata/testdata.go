package testdata

import _ "embed"

const UnknownVin string = "G1YZ23J9P58034278"

//go:embed exampleCar.json
var ExampleCar string

const VinCar string = "WVWAA71K08W201030"

//go:embed exampleCar2.json
var ExampleCar2 string

const VinCar2 string = "1FVNY5Y90HP312888"
