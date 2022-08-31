package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz" // All supported characters.

// Will be called automatically when the package is first used
func init() {
	// Set the seed value for the random generator using Seed() which expects an int64 as input.
	// Normally the Seed value is often set to the current time.
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	// rand.Int63n(max-min+1): a random integer between 0 and max-min.
	// => the final result will be a random integer between min and max.
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder // Declare a new string builder object 'sb'
	k := len(alphabet)     // Get the total number of characters in the alphabets and assign it to k

	for i := 0; i < n; i++ { // Use a single loop to generate n random characters.
		c := alphabet[rand.Intn(k)] // rand.Intn(k): get a random position from 0 to k-1, then take the corresponding char at that pos in the alphabet.
		sb.WriteByte(c)             // Write c to String Builder.
	}

	return sb.String() // Return this to the caller.
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6) // A random string of 6 letters
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000) // A random Int between 0 and 1000
}

// RandomCurrency generates a random currency code
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)] // rand.Intn(n): a random index (0 -> n-1)
}
