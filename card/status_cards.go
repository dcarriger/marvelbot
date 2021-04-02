package card

var Confused = &Card{
	Name: "Confused",
	ImageSrc: "images/confused.png",
}

var Stunned = &Card{
	Name: "Stunned",
	ImageSrc: "images/stunned.png",
}

var Tough = &Card{
	Name: "Tough",
	ImageSrc: "images/tough.png",
}

var StatusCards = []*Card{
	Confused,
	Stunned,
	Tough,
}