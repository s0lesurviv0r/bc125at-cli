package scanner

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Special CTCSS/DCS values
var specialTones = map[int]string{
	0:   "none",
	127: "search",
	240: "no_tone",
}

// CTCSS tones: codes 64-113
var ctcssTones = map[int]string{
	64:  "ctcss_67.0",
	65:  "ctcss_71.9",
	66:  "ctcss_74.4",
	67:  "ctcss_77.0",
	68:  "ctcss_79.7",
	69:  "ctcss_82.5",
	70:  "ctcss_85.4",
	71:  "ctcss_88.5",
	72:  "ctcss_91.5",
	73:  "ctcss_94.8",
	74:  "ctcss_97.4",
	75:  "ctcss_100.0",
	76:  "ctcss_103.5",
	77:  "ctcss_107.2",
	78:  "ctcss_110.9",
	79:  "ctcss_114.8",
	80:  "ctcss_118.8",
	81:  "ctcss_123.0",
	82:  "ctcss_127.3",
	83:  "ctcss_131.8",
	84:  "ctcss_136.5",
	85:  "ctcss_141.3",
	86:  "ctcss_146.2",
	87:  "ctcss_151.4",
	88:  "ctcss_156.7",
	89:  "ctcss_159.8",
	90:  "ctcss_162.2",
	91:  "ctcss_165.5",
	92:  "ctcss_167.9",
	93:  "ctcss_171.3",
	94:  "ctcss_173.8",
	95:  "ctcss_177.3",
	96:  "ctcss_179.9",
	97:  "ctcss_183.5",
	98:  "ctcss_186.2",
	99:  "ctcss_189.9",
	100: "ctcss_192.8",
	101: "ctcss_196.6",
	102: "ctcss_199.5",
	103: "ctcss_203.5",
	104: "ctcss_206.5",
	105: "ctcss_210.7",
	106: "ctcss_218.1",
	107: "ctcss_225.7",
	108: "ctcss_229.1",
	109: "ctcss_233.6",
	110: "ctcss_241.8",
	111: "ctcss_250.3",
	112: "ctcss_254.1",
}

// DCS tones: codes 128-231
var dcsTones = map[int]string{
	128: "dcs_023",
	129: "dcs_025",
	130: "dcs_026",
	131: "dcs_031",
	132: "dcs_032",
	133: "dcs_036",
	134: "dcs_043",
	135: "dcs_047",
	136: "dcs_051",
	137: "dcs_053",
	138: "dcs_054",
	139: "dcs_065",
	140: "dcs_071",
	141: "dcs_072",
	142: "dcs_073",
	143: "dcs_074",
	144: "dcs_114",
	145: "dcs_115",
	146: "dcs_116",
	147: "dcs_122",
	148: "dcs_125",
	149: "dcs_131",
	150: "dcs_132",
	151: "dcs_134",
	152: "dcs_143",
	153: "dcs_145",
	154: "dcs_152",
	155: "dcs_155",
	156: "dcs_156",
	157: "dcs_162",
	158: "dcs_165",
	159: "dcs_172",
	160: "dcs_174",
	161: "dcs_205",
	162: "dcs_212",
	163: "dcs_223",
	164: "dcs_225",
	165: "dcs_226",
	166: "dcs_243",
	167: "dcs_244",
	168: "dcs_245",
	169: "dcs_246",
	170: "dcs_251",
	171: "dcs_252",
	172: "dcs_255",
	173: "dcs_261",
	174: "dcs_263",
	175: "dcs_265",
	176: "dcs_266",
	177: "dcs_271",
	178: "dcs_274",
	179: "dcs_306",
	180: "dcs_311",
	181: "dcs_315",
	182: "dcs_325",
	183: "dcs_331",
	184: "dcs_332",
	185: "dcs_343",
	186: "dcs_346",
	187: "dcs_351",
	188: "dcs_356",
	189: "dcs_364",
	190: "dcs_365",
	191: "dcs_371",
	192: "dcs_411",
	193: "dcs_412",
	194: "dcs_413",
	195: "dcs_423",
	196: "dcs_431",
	197: "dcs_432",
	198: "dcs_445",
	199: "dcs_446",
	200: "dcs_452",
	201: "dcs_454",
	202: "dcs_455",
	203: "dcs_462",
	204: "dcs_464",
	205: "dcs_465",
	206: "dcs_466",
	207: "dcs_503",
	208: "dcs_506",
	209: "dcs_516",
	210: "dcs_523",
	211: "dcs_526",
	212: "dcs_532",
	213: "dcs_546",
	214: "dcs_565",
	215: "dcs_606",
	216: "dcs_612",
	217: "dcs_624",
	218: "dcs_627",
	219: "dcs_631",
	220: "dcs_632",
	221: "dcs_654",
	222: "dcs_662",
	223: "dcs_664",
	224: "dcs_703",
	225: "dcs_712",
	226: "dcs_723",
	227: "dcs_731",
	228: "dcs_732",
	229: "dcs_734",
	230: "dcs_743",
	231: "dcs_754",
}

// ValidTones is the merged map of all valid CTCSS/DCS codes to their human names.
var ValidTones map[int]string

func init() {
	ValidTones = make(map[int]string, len(specialTones)+len(ctcssTones)+len(dcsTones))
	for k, v := range specialTones {
		ValidTones[k] = v
	}
	for k, v := range ctcssTones {
		ValidTones[k] = v
	}
	for k, v := range dcsTones {
		ValidTones[k] = v
	}
}

// ToneToHuman converts an internal CTCSS/DCS code to a human-readable string.
func ToneToHuman(code int) (string, error) {
	if name, ok := ValidTones[code]; ok {
		return name, nil
	}
	return "", fmt.Errorf("invalid CTCSS/DCS code: %d", code)
}

// ToneToInternal converts a human-readable tone string to its internal code.
// Matching is case-insensitive and ignores non-alphanumeric characters.
func ToneToInternal(s string) (int, error) {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	normalized := strings.ToLower(re.ReplaceAllString(s, ""))
	for code, name := range ValidTones {
		normalizedName := strings.ToLower(re.ReplaceAllString(name, ""))
		if normalizedName == normalized {
			return code, nil
		}
	}
	return 0, fmt.Errorf("unknown tone: %q (use 'bc125at tones' to list valid values)", s)
}

// SortedToneKeys returns a sorted slice of all valid tone codes.
func SortedToneKeys() []int {
	keys := make([]int, 0, len(ValidTones))
	for k := range ValidTones {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
