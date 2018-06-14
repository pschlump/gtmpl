package DLLString

type DLLNodeString struct {
	next, prev *DLLNodeString
	data       string
}

type DLLString struct {
	head, tail *DLLNodeString
}

func NewDLLstring() *DLLString {
	return &DLLString{}
}

func (dt *DLLString) InsertHead(data string) {
	nn := &DLLNodeString{
		data: data,
	}
	if dt.head != nil {
		tmp := dt.head
		nn.next = tmp
		tmp.prev = nn
		dt.head = nn
	} else {
		dt.tail = nn
		dt.head = nn
	}
}

// In a real DLL data type you would want to have other fuctions like InsertTail, PopHead, PopTail etc...
