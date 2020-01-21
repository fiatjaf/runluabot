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

func runlua(code string) (res string, err error) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	fmt.Println(code)

	code = fmt.Sprintf(`
sandbox_env = {
  ipairs = ipairs,
  next = next,
  pairs = pairs,
  error = error,
  tonumber = tonumber,
  tostring = tostring,
  type = type,
  unpack = unpack,
  string = { byte = string.byte, char = string.char, find = string.find,
      format = string.format, gmatch = string.gmatch, gsub = string.gsub,
      len = string.len, lower = string.lower, match = string.match,
      rep = string.rep, reverse = string.reverse, sub = string.sub,
      upper = string.upper },
  table = { insert = table.insert, maxn = table.maxn, remove = table.remove,
      sort = table.sort },
  math = { abs = math.abs, acos = math.acos, asin = math.asin,
      atan = math.atan, atan2 = math.atan2, ceil = math.ceil, cos = math.cos,
      cosh = math.cosh, deg = math.deg, exp = math.exp, floor = math.floor,
      fmod = math.fmod, frexp = math.frexp, huge = math.huge,
      ldexp = math.ldexp, log = math.log, log10 = math.log10, max = math.max,
      min = math.min, modf = math.modf, pi = math.pi, pow = math.pow,
      rad = math.rad, random = math.random, randomseed = math.randomseed,
      sin = math.sin, sinh = math.sinh, sqrt = math.sqrt, tan = math.tan, tanh = math.tanh },
  os = { clock = os.clock, difftime = os.difftime, time = os.time, date = os.date },
}

function run (sandbox_env, func, ...)
  local original = _ENV
  _ENV = sandbox_env
  local ret = func(...)
  _ENV = original
  return ret
end

function call ()
  return %s
end

_calls = 0
debug.sethook(function ()
  _calls = _calls + 1
  if _calls > 100 then
    error('timeout!')
  end
end, 'c')

ret = run(sandbox_env, call)
    `, code)

	err = L.DoString(code)
	if err != nil {
		st := stackTraceWithCode(err.Error(), code)
		log.Print(st)
		err = errors.New(st)
		return
	}

	globalsAfter := lunatico.GetGlobals(L, "ret")
	bres, _ := json.Marshal(globalsAfter["ret"])

	return string(bres), nil
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
