package player

import "testing"

func TestGenerateReply(t *testing.T) {
	testMe := func(input, expected string) {
		res := generateReply(input)
		if res != expected {
			t.Error("Payload:", input, "Expected:", expected, "Got:", res)
		}
	}
	testMe("Tikrinu: atrašyk 56", "56")
	testMe("Tikrinu 1925", "esu")
	testMe("tykrinu: 124", "esu")
	testMe("2 Tikrinu: atrašyk 102", "102")
	testMe("atrasik is kyto galo 102", "nebesvaik")
	testMe("atrasyk atvirksciai: labas", "nebesvaik")
	testMe("kiek bus 2+2???", "?")
}
