package support

import (
	"net/http"
	"strconv"
	"strings"
)

func NegotiateContentTypeOf(req *http.Request, offers ...string) string {
	return NegotiateContentType(req.Header.Get("Accept"), offers)
}

func NegotiateContentType(requestedAccept string, offers []string) string {
	accepts := ParseAcceptRanges(requestedAccept)
	var offerRanges []AcceptRange
	for _, off := range offers {
		offerRanges = append(offerRanges, ParseAcceptRange(off))
	}

	return negotiateContentTypeInternal(accepts, offerRanges, offerRanges[len(offerRanges)-1])
}

func negotiateContentTypeInternal(accepts []AcceptRange, offers []AcceptRange, defaultOffer AcceptRange) string {
	best := defaultOffer.RawString()
	bestWeight := float64(0)
	bestParams := 0

	for _, offer := range offers {
		for _, accept := range accepts {
			// add a booster on the weights to prefer more exact matches to wildcards
			// such that: */* = 0, x/* = 1, x/x = 2
			booster := float64(0)
			if accept.Type != "*" {
				booster++
				if accept.Subtype != "*" {
					booster++
				}
			}

			if bestWeight > (accept.Weight + booster) {
				continue // we already have something better..
			} else if accept.Type == "*" && accept.Subtype == "*" {
				best = offer.RawString()
				bestWeight = accept.Weight + booster
			} else if accept.Subtype == "*" && offer.Type == accept.Type {
				best = offer.RawString()
				bestWeight = accept.Weight + booster
			} else if accept.Type == offer.Type && accept.Subtype == offer.Subtype {
				paramCount := compareParams(accept.Parameters, offer.Parameters)
				if paramCount >= bestParams { // if it's equal this one must be better, since the weight was better..
					best = offer.RawString()
					bestWeight = accept.Weight + booster
					bestParams = paramCount
				}
			}
		}
	}

	return best
}

func compareParams(params1 map[string]string, params2 map[string]string) (count int) {
	for k1, v1 := range params1 {
		if v2, ok := params2[k1]; ok && v1 == v2 {
			count++
		}
	}
	return count
}

// AcceptRange represents a media-range contained within an Accept header.
type AcceptRange struct {
	// Type represents the media type.
	Type string
	// Subtype represents the media subtype.
	Subtype string
	// Weight represents the weight (quality factor) of this range.
	Weight float64
	// Parameters represents the parameters that are applicable to this range.
	Parameters map[string]string
	raw        string
}

// RawString returns the raw string for this accept range.
func (a AcceptRange) RawString() string {
	return a.raw
}

// https://tools.ietf.org/html/rfc7231#section-5.3.2
// Accept = #( media-range [ accept-params ] )
//  media-range    = ( "*/*"
//                   / ( type "/" "*" )
//                   / ( type "/" subtype )
//                   ) *( OWS ";" OWS parameter )
//  accept-params  = weight *( accept-ext )
//  accept-ext = OWS ";" OWS token [ "=" ( token / quoted-string ) ]

// ParseAcceptRanges returns the set of accepted media ranges from an Accept header.
func ParseAcceptRanges(accepts string) []AcceptRange {
	result := []AcceptRange{}
	remaining := accepts
	for {
		var accept string
		accept, remaining = extractFieldAndSkipToken(remaining, ',')
		result = append(result, ParseAcceptRange(accept))
		if len(remaining) == 0 {
			break
		}
	}
	return result
}

// ParseAcceptRange returns the media range, params and quality factor (weight) from an Accept range.
func ParseAcceptRange(accept string) AcceptRange {
	typeAndSub, rawparams := extractFieldAndSkipToken(accept, ';')

	tp, subtp := extractFieldAndSkipToken(typeAndSub, '/')
	params := extractParams(rawparams)

	w := extractWeight(params)
	return AcceptRange{Type: tp, Subtype: subtp, Parameters: params, Weight: w, raw: accept}
}

func extractWeight(params map[string]string) float64 {
	if w, ok := params["q"]; ok {
		res, err := strconv.ParseFloat(w, 64)
		if err == nil {
			return res
		}
	}
	return 1 // default is 1
}

func extractParams(raw string) map[string]string {
	params := map[string]string{}
	rest := raw
	for {
		var p string
		p, rest = extractFieldAndSkipToken(rest, ';')
		if len(p) > 0 {
			k, v := extractFieldAndSkipToken(p, '=')
			params[k] = v
		}
		if len(rest) == 0 {
			break
		}
	}

	return params
}

func extractFieldAndSkipToken(s string, sep rune) (string, string) {
	f, r := extractField(s, sep)
	if len(r) > 0 {
		r = r[1:]
	}
	return f, r
}

func extractField(s string, sep rune) (field, rest string) {
	field = s
	for i, v := range s {
		if v == sep {
			field = strings.TrimSpace(s[:i])
			rest = strings.TrimSpace(s[i:])
			break
		}
	}
	return
}
