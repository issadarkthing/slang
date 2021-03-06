package xlisp

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spy16/sabre"
)

// Case implements the switch case construct.
func Case(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	err := checkArityAtLeast(2, len(args))
	if err != nil {
		return nil, err
	}

	res, err := sabre.Eval(scope, args[0])
	if err != nil {
		return nil, err
	}

	if len(args) == 2 {
		return sabre.Eval(scope, args[1])
	}

	start := 1
	for ; start < len(args); start += 2 {
		val := args[start]
		if start+1 >= len(args) {
			return val, nil
		}

		if sabre.Compare(res, val) {
			return sabre.Eval(scope, args[start+1])
		}
	}

	return nil, fmt.Errorf("no matching clause for '%s'", res)
}

// MacroExpand is a wrapper around the sabre MacroExpand function that
// ignores the expanded bool flag.
func MacroExpand(scope sabre.Scope, f sabre.Value) (sabre.Value, error) {
	f, _, err := sabre.MacroExpand(scope, f)
	return f, err
}

// Throw converts args to strings and returns an error with all the strings
// joined.
func Throw(scope sabre.Scope, args ...sabre.Value) error {
	return errors.New(strings.Trim(MakeString(args...).String(), "\""))
}

// Realize realizes a sequence by continuously calling First() and Next()
// until the sequence becomes nil.
func Realize(seq sabre.Seq) *sabre.List {
	var vals []sabre.Value

	for seq != nil {
		v := seq.First()
		if v == nil {
			break
		}
		vals = append(vals, v)
		seq = seq.Next()
	}

	return &sabre.List{Values: vals}
}

// TypeOf returns the type information object for the given argument.
func TypeOf(v interface{}) sabre.Value {
	return sabre.ValueOf(reflect.TypeOf(v))
}

// Implements checks if given value implements the interface represented
// by 't'. Returns error if 't' does not represent an interface type.
func Implements(v interface{}, t sabre.Type) (bool, error) {
	if t.T.Kind() == reflect.Ptr {
		t.T = t.T.Elem()
	}

	if t.T.Kind() != reflect.Interface {
		return false, fmt.Errorf("type '%s' is not an interface type", t)
	}

	return reflect.TypeOf(v).Implements(t.T), nil
}

// ToType attempts to convert given sabre value to target type. Returns
// error if conversion not possible.
func ToType(to sabre.Type, val sabre.Value) (sabre.Value, error) {
	rv := reflect.ValueOf(val)
	if rv.Type().ConvertibleTo(to.T) || rv.Type().AssignableTo(to.T) {
		return sabre.ValueOf(rv.Convert(to.T).Interface()), nil
	}

	return nil, fmt.Errorf("cannot convert '%s' to '%s'", rv.Type(), to.T)
}

// ThreadFirst threads the expressions through forms by inserting result of
// eval as first argument to next expr.
func ThreadFirst(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {
	return threadCall(scope, args, false)
}

// ThreadLast threads the expressions through forms by inserting result of
// eval as last argument to next expr.
func ThreadLast(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {
	return threadCall(scope, args, true)
}

// MakeString returns stringified version of all args.
func MakeString(vals ...sabre.Value) sabre.Value {
	argc := len(vals)
	switch argc {
	case 0:
		return sabre.String("")

	case 1:
		nilVal := sabre.Nil{}
		if vals[0] == nilVal || vals[0] == nil {
			return sabre.String("")
		}

		return sabre.String(strings.Trim(vals[0].String(), "\""))

	default:
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(strings.Trim(v.String(), "\""))
		}
		return sabre.String(sb.String())
	}
}

func threadCall(scope sabre.Scope, args []sabre.Value, last bool) (sabre.Value, error) {

	err := checkArityAtLeast(1, len(args))
	if err != nil {
		return nil, err
	}

	res := args[0]
	// res, err := sabre.Eval(scope, args[0])
	// if err != nil {
	// 	return nil, err
	// }

	for args = args[1:]; len(args) > 0; args = args[1:] {
		form := args[0]

		switch f := form.(type) {
		case *sabre.List:
			if last {
				f.Values = append(f.Values, res)
			} else {
				f.Values = append([]sabre.Value{f.Values[0], res}, f.Values[1:]...)
			}
			res, err = sabre.Eval(scope, f)
			if v, ok := res.(*sabre.List); ok {
				res = v.Cons(sabre.Symbol{Value: "list"})
			}

		case sabre.Invokable:
			res, err = f.Invoke(scope, res)

		default:
			return nil, fmt.Errorf("%s is not invokable", reflect.TypeOf(res))
		}

		if err != nil {
			return nil, err
		}
	}

	if res, ok := res.(*sabre.List); ok {
		return res.Eval(scope)
	}

	return res, nil
}

func isTruthy(v sabre.Value) bool {
	if v == nil || v == (sabre.Nil{}) {
		return false
	}

	if b, ok := v.(sabre.Bool); ok {
		return bool(b)
	}

	return true
}

func slangRange(args ...int) (Any, error) {
	var result []sabre.Value

	switch len(args) {
	case 1:
		result = createRange(0, args[0], 1)
	case 2:
		result = createRange(args[0], args[1], 1)
	case 3:
		result = createRange(args[0], args[1], args[2])
	}

	return &sabre.List{Values: result}, nil
}

func createRange(min, max, step int) []sabre.Value {

	result := make([]sabre.Value, 0, max-min)
	for i := min; i < max; i += step {
		result = append(result, sabre.Int64(i))
	}
	return result
}

func doSeq(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	arg1 := args[0]
	vecs, ok := arg1.(sabre.Vector)
	if !ok {
		return nil, fmt.Errorf("Invalid type")
	}

	coll, err := vecs.Values[1].Eval(scope)
	if err != nil {
		return nil, err
	}

	l, ok := coll.(sabre.Seq)
	if !ok {
		return nil, fmt.Errorf("Invalid type")
	}

	list := Realize(l)

	symbol, ok := vecs.Values[0].(sabre.Symbol)
	if !ok {
		return nil, fmt.Errorf("invalid type; expected symbol")
	}

	var result sabre.Value
	for _, v := range list.Values {
		scope.Bind(symbol.Value, v)
		for _, body := range args[1:] {
			result, err = body.Eval(scope)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// unsafely swap the value. Does not mutate the value rather just swapping
func swap(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	err := checkArity(2, len(args))
	if err != nil {
		return nil, err
	}

	symbol, ok := args[0].(sabre.Symbol)
	if !ok {
		return nil, fmt.Errorf("Expected symbol")
	}

	value, err := args[1].Eval(scope)
	if err != nil {
		return nil, err
	}

	scope.Bind(symbol.Value, value)
	return value, nil
}

// Returns '(recur & expressions) so it'll be recognize by fn.Invoke method as
// tail recursive function
func recur(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	symbol := sabre.Symbol{
		Value: "recur",
	}

	results, err := evalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	results = append([]sabre.Value{symbol}, results...)
	return &sabre.List{Values: results}, nil
}

// Returns string representation of type
func stringTypeOf(v interface{}) string {
	return reflect.TypeOf(v).String()
}

// Evaluate the expressions in another goroutine; returns chan
func future(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	ch := make(chan sabre.Value)

	go func() {

		val, err := args[0].Eval(scope)
		if err != nil {
			panic(err)
		}

		ch <- val
		close(ch)
	}()

	return sabre.ValueOf(ch), nil
}

// Deref chan from future to get the value. This call is blocking until future is resolved.
// The result will be cached.
func deref(scope sabre.Scope) func(sabre.Symbol, <-chan sabre.Value) (sabre.Value, error) {

	return func(symbol sabre.Symbol, ch <-chan sabre.Value) (sabre.Value, error) {

		derefSymbol := fmt.Sprintf("__deref__%s__result__", symbol.Value)

		value, ok := <-ch
		if ok {
			scope.Bind(derefSymbol, value)
			return value, nil
		}

		value, err := scope.Resolve(derefSymbol)
		if err != nil {
			return nil, err
		}

		return value, nil
	}
}

func sleep(s int) {
	time.Sleep(time.Millisecond * time.Duration(s))
}

func futureRealize(ch <-chan sabre.Value) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func xlispTime(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	var lastVal sabre.Value
	var err error
	initial := time.Now()
	for _, v := range args {
		lastVal, err = v.Eval(scope)
		if err != nil {
			return nil, err
		}
	}
	final := time.Since(initial)
	fmt.Printf("Elapsed time: %s\n", final.String())

	return lastVal, nil
}

func parseLoop(scope sabre.Scope, args []sabre.Value) (*sabre.Fn, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("call requires at-least bindings argument")
	}

	vec, isVector := args[0].(sabre.Vector)
	if !isVector {
		return nil, fmt.Errorf(
			"first argument to let must be bindings vector, not %v",
			reflect.TypeOf(args[0]),
		)
	}

	if len(vec.Values)%2 != 0 {
		return nil, fmt.Errorf("bindings must contain event forms")
	}

	var bindings []binding
	for i := 0; i < len(vec.Values); i += 2 {
		sym, isSymbol := vec.Values[i].(sabre.Symbol)
		if !isSymbol {
			return nil, fmt.Errorf(
				"item at %d must be symbol, not %s",
				i, vec.Values[i],
			)
		}

		bindings = append(bindings, binding{
			Name: sym.Value,
			Expr: vec.Values[i+1],
		})
	}

	return &sabre.Fn{
		Func: func(scope sabre.Scope, _ []sabre.Value) (sabre.Value, error) {
			letScope := sabre.NewScope(scope)
			for _, b := range bindings {
				v, err := b.Expr.Eval(letScope)
				if err != nil {
					return nil, err
				}
				_ = letScope.Bind(b.Name, v)
			}

			result, err := sabre.Module(args[1:]).Eval(letScope)
			if err != nil {
				return nil, err
			}

			for isRecur(result) {

				newBindings := result.(*sabre.List).Values[1:]
				for i, b := range bindings {
					letScope.Bind(b.Name, newBindings[i])
				}

				result, err = sabre.Module(args[1:]).Eval(letScope)
				if err != nil {
					return nil, err
				}
			}

			return result, err
		},
	}, nil
}

func isRecur(value sabre.Value) bool {

	list, ok := value.(*sabre.List)
	if !ok {
		return false
	}

	sym, ok := list.First().(sabre.Symbol)
	if !ok {
		return false
	}

	if sym.Value != "recur" {
		return false
	}

	return true
}

type binding struct {
	Name string
	Expr sabre.Value
}

func and(x sabre.Value, y sabre.Value) bool {
	return isTruthy(x) && isTruthy(y)
}

func or(x sabre.Value, y sabre.Value) bool {
	return isTruthy(x) || isTruthy(y)
}

func safeSwap(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	atom := args[0]
	atom, err := atom.Eval(scope)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve symbol")
	}

	fn, err := args[1].Eval(scope)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve symbol")
	}

	return atom.(*Atom).UpdateState(scope, fn.(sabre.Invokable))
}

func bound(scope sabre.Scope) func(sabre.Symbol) bool {
	return func(sym sabre.Symbol) bool {
		_, err := scope.Resolve(sym.Value)
		return err == nil
	}
}

func resolve(scope sabre.Scope) func(sabre.Symbol) sabre.Value {
	return func(sym sabre.Symbol) sabre.Value {
		val, _ := scope.Resolve(sym.Value)
		return val
	}
}
