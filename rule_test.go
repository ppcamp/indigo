package indigo_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ezachrisen/indigo"
	"github.com/matryer/is"
)

// Test that all rules are evaluated and yield the correct result in the default configuration
func TestEvaluationTraversalDefault(t *testing.T) {
	is := is.New(t)

	m := newMockEvaluator()
	r := makeRule()

	//fmt.Println(r)
	expectedResults := map[string]bool{
		"rule1": true,
		"D":     true,
		"d1":    true,
		"d2":    false,
		"d3":    true,
		"B":     false,
		"b1":    true,
		"b2":    false,
		"b3":    true,
		"b4":    false,
		"b4-1":  true,
		"b4-2":  false,
		"E":     false,
		"e1":    true,
		"e2":    false,
		"e3":    true,
	}

	err := r.Compile(m)
	is.NoErr(err)

	result, err := r.Evaluate(m, map[string]interface{}{})
	is.NoErr(err)
	//fmt.Println(m.rulesTested)
	//fmt.Println(result)
	is.NoErr(match(flattenResults(result), expectedResults))
}

// Ensure that rules are evaluated in the correct order when
// alpha sort is applied
func TestEvaluationTraversalAlphaSort(t *testing.T) {
	is := is.New(t)

	e := newMockEvaluator()
	r := makeRule()

	// Specify the sort order for all rules
	apply(r, func(r *indigo.Rule) error {
		r.SortFunc = sortRulesAlpha
		return nil
	})

	err := r.Compile(e)
	is.NoErr(err)

	//	fmt.Println(r)
	expectedResults := map[string]bool{
		"rule1": true,
		"D":     true,
		"d1":    true,
		"d2":    false,
		"d3":    true,
		"B":     false,
		"b1":    true,
		"b2":    false,
		"b3":    true,
		"b4":    false,
		"b4-1":  true,
		"b4-2":  false,
		"E":     false,
		"e1":    true,
		"e2":    false,
		"e3":    true,
	}

	//If everything works, the rules were evaluated in this order
	//	(alphabetically)
	expectedOrder := []string{
		"rule1",
		"B",
		"b1",
		"b2",
		"b3",
		"b4",
		"b4-1",
		"b4-2",
		"D",
		"d1",
		"d2",
		"d3",
		"E",
		"e1",
		"e2",
		"e3",
	}

	result, err := r.Evaluate(e, map[string]interface{}{})
	is.NoErr(err)
	//	fmt.Println(m.rulesTested)
	//fmt.Println(result)
	is.NoErr(match(flattenResults(result), expectedResults))
	is.True(reflect.DeepEqual(expectedOrder, e.rulesTested)) // not all rules were evaluated
}

// Test that a self reference is passed through compilation, evaluation
// and finally returned in results
func TestSelf(t *testing.T) {
	is := is.New(t)

	e := newMockEvaluator()
	r := makeRule()
	err := r.Compile(e)

	// Set the self reference on D
	D := r.Rules["D"]
	D.Self = 22
	D.Expr = "self"
	d1 := D.Rules["d1"]
	// Give d1 a self expression, but no self value
	d1.Expr = "self"

	result, err := r.Evaluate(e, nil)
	is.True(err != nil) // should get an error if the data map is nil and we try to use 'self'

	result, err = r.Evaluate(e, map[string]interface{}{"anything": "anything"})
	is.NoErr(err)
	is.Equal(result.Results["D"].Value.(int), 22)           // D should return 'self', which is 22
	is.Equal(result.Results["D"].Results["d1"].Pass, false) // d1 should not inherit D's self
}

// Test that the engine checks for nil data and rule
func TestNilDataOrRule(t *testing.T) {
	is := is.New(t)

	e := newMockEvaluator()
	r := makeRule()

	_, err := r.Evaluate(e, nil)
	is.True(err != nil) // should get an error if the data map is nil
	is.True(strings.Contains(err.Error(), "data is nil"))

	_, err = r.Evaluate(nil, map[string]interface{}{})
	is.True(err != nil) // should get an error if the rule is nil
	is.True(strings.Contains(err.Error(), "evaluator is nil"))

	r.Rules["B"].Rules["oops"] = nil
	_, err = r.Evaluate(e, map[string]interface{}{})
	is.True(err != nil) // should get an error if the rule is nil
	is.True(strings.Contains(err.Error(), "rule is nil"))

}

// Test the rule evaluation with various options set on the rules
func TestEvalOptions(t *testing.T) {
	is := is.New(t)

	e := newMockEvaluator()
	d := map[string]interface{}{"a": "a"} // dummy data, not important
	w := map[string]bool{                 // the wanted rule evaluation results with no options in effect
		"rule1": true,
		"D":     true,
		"d1":    true,
		"d2":    false,
		"d3":    true,
		"B":     false,
		"b1":    true,
		"b2":    false,
		"b3":    true,
		"b4":    false,
		"b4-1":  true,
		"b4-2":  false,
		"E":     false,
		"e1":    true,
		"e2":    false,
		"e3":    true,
	}

	cases := map[string]struct {
		prep func(*indigo.Rule)     // a function called to prep the rule before compilation and evaluation
		want func() map[string]bool // a function returning an edited copy of w after options are applied
	}{
		"Default Options": {
			prep: func(r *indigo.Rule) {
			},
			want: func() map[string]bool {
				return copyMap(w)
			},
		},

		"StopIfParentNegative": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].StopIfParentNegative = true
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b1", "b2", "b3", "b4", "b4-1", "b4-2")
			},
		},
		"DiscardPass": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].Rules["b4"].DiscardPass = true
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b4-1")
			},
		},
		"DiscardFail": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].DiscardFail = true
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b2", "b4", "b4-1", "b4-2")
			},
		},
		"DiscardPass & DiscardFail": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].DiscardFail = true
				r.Rules["B"].DiscardPass = true
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b1", "b2", "b3", "b4", "b4-1", "b4-2")
			},
		},
		"DiscardPass & DiscardFail on Root": {
			prep: func(r *indigo.Rule) {
				r.DiscardFail = true
				r.DiscardPass = true
			},
			want: func() map[string]bool {
				return map[string]bool{"rule1": true}
			},
		},
		"StopFirstPositiveChild": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].StopFirstPositiveChild = true
				r.Rules["B"].SortFunc = sortRulesAlpha
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b2", "b3", "b4", "b4-1", "b4-2")
			},
		},
		"StopFirstNegativeChild": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].StopFirstNegativeChild = true
				r.Rules["B"].SortFunc = sortRulesAlpha
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b3", "b4", "b4-1", "b4-2")
			},
		},
		"StopFirstNegativeChild & StopFirstPositiveChild": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].StopFirstNegativeChild = true
				r.Rules["B"].StopFirstPositiveChild = true
				r.Rules["B"].SortFunc = sortRulesAlpha
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b2", "b3", "b4", "b4-1", "b4-2")
			},
		},
		"Multiple Options": {
			prep: func(r *indigo.Rule) {
				r.Rules["B"].DiscardPass = true
				r.Rules["B"].Rules["b4"].DiscardPass = true
				r.Rules["E"].StopIfParentNegative = true
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b1", "b3", "b4-1", "e1", "e2", "e3")
			},
		},
		"Delete Rule": {
			prep: func(r *indigo.Rule) {
				delete(r.Rules["B"].Rules, "b1")
			},
			want: func() map[string]bool {
				return deleteKeys(copyMap(w), "b1")
			},
		},
		"Add Rule": {
			prep: func(r *indigo.Rule) {
				x := &indigo.Rule{
					ID:   "x",
					Expr: "true",
				}
				r.Rules[x.ID] = x
			},
			want: func() map[string]bool {
				w := copyMap(w)
				w["x"] = true
				return w
			},
		},
		"Update Rule": { // TODO and forget to compile; stale program?
			prep: func(r *indigo.Rule) {
				r.Rules["B"].Rules["b4"].Expr = "true"
			},
			want: func() map[string]bool {
				w := copyMap(w)
				w["b4"] = true
				return w
			},
		},
	}

	for k, c := range cases {
		e.ResetRulesTested()
		r := makeRule()
		c.prep(r)

		u, err := r.Evaluate(e, d)
		is.NoErr(err)

		err = match(flattenResults(u), c.want())
		if err != nil {
			t.Errorf("Error in case %s: %v", k, err)
			fmt.Println(r)
			fmt.Println(u)
		}
	}
}

func TestDiagnosticOptions(t *testing.T) {

	is := is.New(t)

	e := newMockEvaluator()
	d := map[string]interface{}{"a": "a"} // dummy data, not important

	cases := map[string]struct {
		engineDiagnosticCompileRequired bool // whether engine should require compile-time diagnostics
		compileDiagnostics              bool // whether to request diagnostics at compile time
		evalDiagnostics                 bool // whether to request diagnostics at eval time
		wantDiagnostics                 bool // whether we expect to received diagnostic information
	}{
		"No diagnostics requested at compile or eval time": {
			engineDiagnosticCompileRequired: true,
			compileDiagnostics:              false,
			evalDiagnostics:                 false,
			wantDiagnostics:                 false,
		},
		"No diagnostics requested at compile time, but NOT at eval time": {
			engineDiagnosticCompileRequired: true,
			compileDiagnostics:              false,
			evalDiagnostics:                 true,
			wantDiagnostics:                 false,
		},
		"Diagnostics requested at compile time AND at eval time": {
			engineDiagnosticCompileRequired: true,
			compileDiagnostics:              true,
			evalDiagnostics:                 true,
			wantDiagnostics:                 true,
		},
	}

	for k, c := range cases {
		e.ResetRulesTested()
		r := makeRule()

		// Set the mock engine to require that diagnostics must be turned on at compile time,
		// or not. This is a special feature of the mock engine, useful for testing.
		e.diagnosticCompileRequired = c.engineDiagnosticCompileRequired

		err := r.Compile(e, indigo.CollectDiagnostics(c.compileDiagnostics))
		is.NoErr(err)

		u, err := r.Evaluate(e, d, indigo.ReturnDiagnostics(c.evalDiagnostics))
		is.NoErr(err)

		switch c.wantDiagnostics {
		case true:
			err = allNotEmpty(flattenResultsDiagnostics(u))
			if err != nil {
				t.Errorf("In case '%s', wanted diagnostics: %v", k, err)
			}
		default:
			err = anyNotEmpty(flattenResultsDiagnostics(u))
			if err != nil {
				t.Errorf("In case '%s', wanted no diagnostics, got: %v", k, err)
			}
		}

	}

}

// Test compiling some rules with compile-time collection of diagnostics and some without
func TestPartialDiagnostics(t *testing.T) {

	is := is.New(t)

	e := newMockEvaluator()
	// Set the mock engine to require that diagnostics must be turned on at compile time,
	// or not. This is a special feature of the mock engine, useful for testing.
	e.diagnosticCompileRequired = true
	d := map[string]interface{}{"a": "a"} // dummy data, not important
	r := makeRule()

	// first, compile all the rules WITHOUT diagnostics
	err := r.Compile(e)
	is.NoErr(err)

	// then, re-compile rule B, WITH diagnostics
	err = r.Rules["B"].Compile(e, indigo.CollectDiagnostics(true))

	u, err := r.Evaluate(e, d, indigo.ReturnDiagnostics(true))
	is.NoErr(err)

	f := flattenResultsDiagnostics(u)

	// We expect that only the B rules will have diagnostics
	for k, v := range f {
		if k == "B" || k == "b1" || k == "b2" || k == "b3" || k == "b4" || k == "b4-1" || k == "b4-2" {
			is.True(v == "diagnostics here")
		} else {
			if v != "" {
				fmt.Println("V = ", v, k)
			}
			is.True(v == "")
		}
	}
}