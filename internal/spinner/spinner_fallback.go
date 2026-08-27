//go:build !windows

package spinner

// Non-windows platforms always support unicode natively
var SupportsBrailleCharacters = true
