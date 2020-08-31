package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/lunatico"
)

func runlua(actualCode string, pretty bool) (res string, err error) {
	actualCode = strings.TrimSpace(actualCode)
	if actualCode == "" {
		return "", nil
	}

	if strings.HasPrefix(actualCode, "function ") ||
		strings.HasPrefix(actualCode, "for ") ||
		strings.HasPrefix(actualCode, "local ") ||
		strings.HasPrefix(actualCode, "repeat") ||
		strings.HasPrefix(actualCode, ";") ||
		strings.HasPrefix(actualCode, "if ") {

	} else {
		actualCode = "return " + actualCode
	}

	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	lunatico.SetGlobals(L, map[string]interface{}{
		"code": actualCode,
	})

	code := `
sandbox_env = {
  ipairs = ipairs,
  next = next,
  pairs = pairs,
  error = error,
  tonumber = tonumber,
  tostring = tostring,
  type = type,
  unpack = unpack,
  utf8 = utf8,
  string = { byte = string.byte, char = string.char, find = string.find,
      format = string.format, gmatch = string.gmatch, gsub = string.gsub,
      len = string.len, lower = string.lower, match = string.match,
      rep = string.rep, reverse = string.reverse, sub = string.sub,
      upper = string.upper },
  table = { insert = table.insert, maxn = table.maxn, remove = table.remove,
      sort = table.sort, pack = table.pack },
  math = { abs = math.abs, acos = math.acos, asin = math.asin,
      atan = math.atan, atan2 = math.atan2, ceil = math.ceil, cos = math.cos,
      cosh = math.cosh, deg = math.deg, exp = math.exp, floor = math.floor,
      fmod = math.fmod, frexp = math.frexp, huge = math.huge,
      ldexp = math.ldexp, log = math.log, log10 = math.log10, max = math.max,
      min = math.min, modf = math.modf, pi = math.pi, pow = math.pow,
      rad = math.rad, random = math.random, randomseed = math.randomseed,
      sin = math.sin, sinh = math.sinh, sqrt = math.sqrt, tan = math.tan, tanh = math.tanh },
  os = { clock = os.clock, difftime = os.difftime, time = os.time, date = os.date },
  print = function (...)
    local args = table.pack(...)
    printed[#printed + 1] = {}
    for i, v in ipairs(args) do
      printed[#printed][i] = tostring(v)
    end
  end,
}

printed = {}

_calls = 0
function count ()
  _calls = _calls + 1
  if _calls > 100 then
    error('timeout!')
  end
end
debug.sethook(count, 'c')

ret = load(code, 'runenv', 't', sandbox_env)()
    `
	err = L.DoString(code)
	if err != nil {
		st := stackTraceWithCode(err.Error(), actualCode)
		log.Print(st)
		err = errors.New(st)
		return
	}

	var result string

	globalsAfter := lunatico.GetGlobals(L, "ret", "printed")

	printed := globalsAfter["printed"]
	if printedarr, ok := printed.([]interface{}); ok {
		for _, line := range printedarr {
			if rows, ok := line.([]interface{}); ok {
				for _, row := range rows {
					result += row.(string) + "\t"
				}
				result += "\n"
			}
		}
	}

	format := json.Marshal
	if pretty {
		format = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		}
	}

	bret, _ := format(globalsAfter["ret"])
	ret := string(bret)
	if result == "" || ret != "null" {
		result += ret
	}
	log.Debug().Str("code", actualCode).Str("result", result).Msg("ran")

	return result, nil
}

var reNumber = regexp.MustCompile("\\d+")

func stackTraceWithCode(stacktrace string, code string) string {
	var result []string

	stlines := strings.Split(stacktrace, "\n")
	lines := strings.Split(code, "\n")
	// result = append(result, stlines[0])

	for i := 0; i < len(stlines); i++ {
		stline := stlines[i]
		result = append(result, stline)

		snum := reNumber.FindString(stline)
		if snum != "" {
			num, _ := strconv.Atoi(snum)
			for i, line := range lines {
				line = fmt.Sprintf("%3d %s", i+1, line)
				if i+1 > num-3 && i+1 < num+3 {
					result = append(result, line)
				}
			}
		}
	}

	return strings.Join(result, "\n")
}
