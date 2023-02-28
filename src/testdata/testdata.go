package testdata

import _ "embed"

const UnknownVin string = "G1YZ23J9P58034278"

//go:embed exampleCar.json
var ExampleCar string

const VinCar string = "WVWAA71K08W201030"

//go:embed exampleCar2.json
var ExampleCar2 string

//go:embed exampleCarStaticResponse.json
var ExampleCarStaticResponse string

const VinCar2 string = "1FVNY5Y90HP312888"

//go:embed locked.json
var Locked string

//go:embed unlocked.json
var Unlocked string

const TrunkAccessToken = "bumrLuCMbumrLuCMbumrLuCM"

//go:embed carsAvailableFirst.json
var CarsAvailableFirst string

//go:embed carsAvailableBoth.json
var CarsAvailableBoth string

//go:embed customerRentalUpcoming.json
var CustomerRentalUpcoming string

//go:embed customerRentalExpired.json
var CustomerRentalExpired string

//go:embed customerRentalActive.json
var CustomerRentalActive string

//go:embed exampleCarVins.json
var ExampleCarVins string

var EmptyArray = "[]"

//go:embed timePeriod1900.json
var TimePeriod1900 string

//go:embed timePeriod2122.json
var TimePeriod2122 string

//go:embed timePeriod2122-23.json
var TimePeriod2122To23 string

//go:embed timePeriod2123.json
var TimePeriod2123 string

//go:embed timePeriod2150.json
var TimePeriod2150 string

//go:embed timePeriodLong.json
var TimePeriodLong string

//go:embed timePeriodSemanticInvalid.json
var TimePeriodSemanticInvalid string

//go:embed timePeriodSyntaxInvalid.json
var TimePeriodSyntaxInvalid string
