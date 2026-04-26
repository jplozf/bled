package utils

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"math"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type HSL struct {
	H, S, L float64
}

// ****************************************************************************
// GLOBALS
// ****************************************************************************
var TargetHues = map[string]float64{
	"Red":     0,
	"Orange":  30,
	"Yellow":  60,
	"Green":   120,
	"Cyan":    180,
	"Bleu":    240,
	"Purple":  270,
	"Magenta": 300,
}

// ****************************************************************************
// GetClosestTargetHue()
// ****************************************************************************
func GetClosestTargetHue(r, g, b uint8) (string, float64) {
	hsl := RGBToHSL(r, g, b)
	inputHue := hsl.H

	bestName := ""
	minDistance := 360.0

	for name, targetHue := range TargetHues {
		dist := math.Abs(inputHue - targetHue)
		if dist > 180 {
			dist = 360 - dist
		}

		if dist < minDistance {
			minDistance = dist
			bestName = name
		}
	}
	return bestName, TargetHues[bestName]
}

// ****************************************************************************
// RGBToHSL()
// ****************************************************************************
func RGBToHSL(r, g, b uint8) HSL {
	R, G, B := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0

	max := math.Max(R, math.Max(G, B))
	min := math.Min(R, math.Min(G, B))
	delta := max - min

	var h, s, l float64
	l = (max + min) / 2

	if delta == 0 {
		h = 0
		s = 0
	} else {
		if l < 0.5 {
			s = delta / (max + min)
		} else {
			s = delta / (2.0 - max - min)
		}

		switch max {
		case R:
			h = (G - B) / delta
			if G < B {
				h += 6
			}
		case G:
			h = (B-R)/delta + 2
		case B:
			h = (R-G)/delta + 4
		}
		h /= 6
	}

	return HSL{H: h * 360, S: s, L: l}
}

// ****************************************************************************
// GetStyledTextColor()
// ****************************************************************************
// GetStyledTextColor generates a text color in Hex format based on the background color (in RGB) and a target hue.
func GetStyledTextColor(bgR, bgG, bgB uint8, targetHue float64) string {
	bgHSL := RGBToHSL(bgR, bgG, bgB)
	var textL float64
	if bgHSL.L > 0.5 {
		textL = 0.15
	} else {
		textL = 0.85
	}
	r, g, b := HSLToRGB(targetHue, 0.7, textL)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// ****************************************************************************
// HexToRGB()
// ****************************************************************************
func HexToRGB(hex string) (uint8, uint8, uint8) {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// ****************************************************************************
// HexToRGB()
// ****************************************************************************
func HSLToRGB(h, s, l float64) (uint8, uint8, uint8) {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l // Gris
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h/360+1.0/3.0)
		g = hueToRGB(p, q, h/360)
		b = hueToRGB(p, q, h/360-1.0/3.0)
	}

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

// ****************************************************************************
// hueToRGB()
// ****************************************************************************
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// ****************************************************************************
// GetContrastedTextColor()
// ****************************************************************************
func GetContrastedTextColor(hexInput string) string {
	var r, g, b uint8

	if len(hexInput) == 4 {
		fmt.Sscanf(hexInput, "#%1x%1x%1x", &r, &g, &b)
		r *= 17
		g *= 17
		b *= 17
	} else {
		fmt.Sscanf(hexInput, "#%02x%02x%02x", &r, &g, &b)
	}

	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))

	// Threshold for deciding text color, usually around 128, but adjusted for better contrast
	if luminance > 110 {
		return "#000000"
	}
	return "#FFFFFF"
}
