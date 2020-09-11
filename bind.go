package slang

import (
	"math"

	"github.com/spy16/sabre"
)

// BindAll binds all core functions into the given scope.
func BindAll(scope sabre.Scope) error {
	core := map[string]sabre.Value{
		// built-in
		"core/range": sabre.ValueOf(slangRange),

		"core/doseq": &sabre.Fn{
			Args:     []string{"vector", "exprs"},
			Variadic: true,
			Func:     doSeq,
		},

		"unsafe/mutate": &sabre.Fn{
			Args: []string{"vector", "exprs"},
			Func: mutate,
		},

		"core/->": &sabre.Fn{
			Args:     []string{"exprs"},
			Func:     ThreadFirst,
			Variadic: true,
		},
		"core/->>": &sabre.Fn{
			Args:     []string{"exprs"},
			Func:     ThreadLast,
			Variadic: true,
		},
		"core/case": &sabre.Fn{
			Args:     []string{"exprs", "clauses"},
			Func:     Case,
			Variadic: true,
		},

		// special forms
		"core/do":           sabre.Do,
		"core/def":          sabre.Def,
		"core/if":           sabre.If,
		"core/fn*":          sabre.Lambda,
		"core/macro*":       sabre.Macro,
		"core/let*":         sabre.Let,
		"core/quote":        sabre.SimpleQuote,
		"core/syntax-quote": sabre.SyntaxQuote,

		"core/macroexpand": sabre.ValueOf(MacroExpand),
		"core/eval":        sabre.ValueOf(sabre.Eval),
		"core/type":        sabre.ValueOf(TypeOf),
		"core/to-type":     sabre.ValueOf(ToType),
		"core/impl?":       sabre.ValueOf(Implements),
		"core/realize":     sabre.ValueOf(Realize),
		"core/throw":       sabre.ValueOf(Throw),

		// Type system functions
		"core/str": sabre.ValueOf(MakeString),

		// Math functions
		"core/+":   sabre.ValueOf(Add),
		"core/-":   sabre.ValueOf(Sub),
		"core/*":   sabre.ValueOf(Multiply),
		"core//":   sabre.ValueOf(Divide),
		"core/mod": sabre.ValueOf(math.Mod),
		"core/=":   sabre.ValueOf(sabre.Compare),
		"core/>":   sabre.ValueOf(Gt),
		"core/>=":  sabre.ValueOf(GtE),
		"core/<":   sabre.ValueOf(Lt),
		"core/<=":  sabre.ValueOf(LtE),

		// io functions
		"core/print":  sabre.ValueOf(Println),
		"core/printf": sabre.ValueOf(Printf),

		"types/Seq":       TypeOf((*sabre.Seq)(nil)),
		"types/Invokable": TypeOf((*sabre.Invokable)(nil)),
	}

	for sym, val := range core {
		if err := scope.Bind(sym, val); err != nil {
			return err
		}
	}

	return nil
}
