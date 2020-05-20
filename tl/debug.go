package tl

var DbOn map[string]bool

func init() {
	DbOn = make(map[string]bool)
}

func SetDbOn(x map[string]bool) {
	DbOn = x
}
