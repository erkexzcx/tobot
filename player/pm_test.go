package player

import "testing"

func TestGenerateReply(t *testing.T) {
	testMe := func(input string, expectIgnore bool, expected string) {
		res, ignore := generateReply(input)
		if expectIgnore != ignore {
			t.Error("Payload:", input, "Expected:", expectIgnore, "Got:", ignore)
		}
		if res != expected {
			t.Error("Payload:", input, "Expected:", expected, "Got:", res, ignore)
		}
	}
	testMe("Tikrinu: atrašyk 56", false, "56")
	testMe("Tikrinu 1925", false, "?")
	testMe("tykrinu: 124", false, "?")
	testMe("2 Tikrinu: atrašyk 102", false, "102")
	testMe("atrasik is kyto galo 102", false, "nesvaik")
	testMe("atrasyk atvirksciai: labas", false, "nesvaik")
	testMe("kiek bus 2+2???", false, "nustok klausinet")
}
