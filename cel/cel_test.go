package cel_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ezachrisen/indigo"
	"github.com/ezachrisen/indigo/cel"
	"github.com/ezachrisen/indigo/testdata/school"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/cel-go/common/types/pb"
	"github.com/matryer/is"
	// to enable easy checks for map membership
)

func makeStudentData() map[string]interface{} {
	return map[string]interface{}{
		"student.ID":             "12312",
		"student.Age":            16,
		"student.GPA":            2.2,
		"student.Status":         "Enrolled",
		"student.Grades":         []interface{}{"A", "B", "A"},
		"student.EnrollmentDate": "2018-08-03T16:00:00-07:00",
		"student.Adjustment":     2.1,
		"now":                    "2019-08-03T16:00:00-07:00",
		"specificTime":           &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}

}

func makeEducationSchema() indigo.Schema {
	return indigo.Schema{
		Elements: []indigo.DataElement{
			{Name: "student.ID", Type: indigo.String{}},
			{Name: "student.Age", Type: indigo.Int{}},
			{Name: "student.GPA", Type: indigo.Float{}},
			{Name: "student.Adjustment", Type: indigo.Float{}},
			{Name: "student.Status", Type: indigo.String{}},
			{Name: "student.Grades", Type: indigo.List{ValueType: indigo.String{}}},
			{Name: "student.EnrollmentDate", Type: indigo.String{}},
			{Name: "now", Type: indigo.String{}},
			{Name: "alsoNow", Type: indigo.Timestamp{}},
		},
	}

}

func makeEducationRules() []indigo.Rule {
	rule1 := indigo.Rule{
		ID:     "student_actions",
		Meta:   "d04ab6d9-f59d-9474-5c38-34d65380c612",
		Schema: makeEducationSchema(),
		Rules: map[string]indigo.Rule{
			"a": {
				ID:   "honors_student",
				Expr: `student.GPA >= 3.6 && student.Status!="Probation" && !("C" in student.Grades)`,
			},
			"b": {
				ID:   "at_risk",
				Expr: `student.GPA < 2.5 || student.Status == "Probation"`,
				Rules: map[string]indigo.Rule{
					"c": {
						ID:   "risk_factor",
						Expr: `2.0+6.0`,
					},
				},
			},
		},
	}

	rule2 := indigo.Rule{
		ID:     "depthRules",
		Schema: makeEducationSchema(),
		Expr:   `student.GPA > 3.5`, // false
		Rules: map[string]indigo.Rule{
			"a": {
				ID:   "c1",
				Expr: `student.Adjustment > 0.0`, // true
			},
			"b": {
				ID:   "c2",
				Expr: `student.Adjustment > 3.0`, // false
			},
			"c": {
				ID:   "c3",
				Expr: `student.Adjustment < 2.6`, // true
			},
			"d": {
				ID:   "c4",
				Expr: `student.Adjustment > 3.0`, // false
			},
		},
	}

	rule3 := indigo.Rule{
		ID:     "ruleOptions",
		Schema: makeEducationSchema(),
		Expr:   `student.GPA > 3.5`, // false
		Rules: map[string]indigo.Rule{
			"A": {
				ID:       "D",
				Expr:     `student.Adjustment > 0.0`,                               // true
				EvalOpts: []indigo.EvalOption{indigo.StopFirstPositiveChild(true)}, // RULE OPTION
				Rules: map[string]indigo.Rule{
					"d1": {
						ID:   "d1",
						Expr: `student.Adjustment < 2.6`, // true
					},
					"d2": {
						ID:   "d2",
						Expr: `student.Adjustment > 3.0`, // false
					},
					"d3": {
						ID:   "d3",
						Expr: `student.Adjustment < 2.6`, // true
					},
				},
			},
			"B": {
				ID:   "b1",
				Expr: `student.Adjustment > 3.0`, // false
			},
			"E": {
				ID:       "E",
				Expr:     `student.Adjustment > 0.0`, // true
				EvalOpts: []indigo.EvalOption{},      // NO RULE OPTION
				Rules: map[string]indigo.Rule{
					"e1": {
						ID:   "e1",
						Expr: `student.Adjustment < 2.6`, // true
					},
					"e2": {
						ID:   "e2",
						Expr: `student.Adjustment > 3.0`, // false
					},
					"e3": {
						ID:   "e3",
						Expr: `student.Adjustment < 2.6`, // true
					},
				},
			},
		},
	}

	return []indigo.Rule{rule1, rule2, rule3}

}

func TestBasicRules(t *testing.T) {

	is := is.New(t)

	engine := cel.NewEngine()
	rule := makeEducationRules()

	err := engine.AddRule(rule...)
	is.NoErr(err)

	results, err := engine.Evaluate(makeStudentData(), "student_actions")

	//	indigo.PrintResults(results)

	is.NoErr(err)
	is.Equal(results.Meta, rule[0].Meta)
	is.True(results.Pass)
	is.True(!results.Results["honors_student"].Pass)
	is.True(results.Results["at_risk"].Pass)
	is.Equal(results.Results["at_risk"].Results["risk_factor"].Value.(float64), 8.0)

	indigo.PrintResults(results)

}

func TestCalculation(t *testing.T) {

	is := is.New(t)
	engine := cel.NewEngine()
	v, err := engine.Calculate(makeStudentData(), `2.0+student.GPA + (1.344 * student.Adjustment)/3.3`, makeEducationSchema())

	is.NoErr(err)
	is.Equal(v, 5.055272727272728)
}

func makeEducationProtoSchema() indigo.Schema {
	return indigo.Schema{
		Elements: []indigo.DataElement{
			{Name: "student", Type: indigo.Proto{Protoname: "school.Student", Message: &school.Student{}}},
			{Name: "now", Type: indigo.Timestamp{}},
			{Name: "self", Type: indigo.Proto{Protoname: "school.HonorsConfiguration", Message: &school.HonorsConfiguration{}}},
		},
	}
}

func makeEducationProtoRules() indigo.Rule {
	return indigo.Rule{
		ID:     "student_actions",
		Schema: makeEducationProtoSchema(),
		Rules: map[string]indigo.Rule{
			"a": {
				ID:   "honor_student",
				Expr: `student.GPA >= self.Minimum_GPA && student.Status != school.Student.status_type.PROBATION && student.Grades.all(g, g>=3.0)`,
				Self: &school.HonorsConfiguration{Minimum_GPA: 3.7},
				Meta: true,
			},
			"b": {
				ID:   "at_risk",
				Expr: `student.GPA < 2.5 || student.Status == school.Student.status_type.PROBATION`,
				Meta: false,
			},
			"c": {
				ID:   "tenure_gt_6months",
				Expr: `now - student.EnrollmentDate > duration("4320h")`, // 6 months = 4320 hours
				Meta: true,
			},
		},
	}

}

func makeStudentProtoData() map[string]interface{} {
	s := school.Student{
		Age:            16,
		GPA:            3.76,
		Status:         school.Student_ENROLLED,
		Grades:         []float64{4.0, 4.0, 3.7},
		Attrs:          map[string]string{"Nickname": "Joey"},
		EnrollmentDate: &timestamp.Timestamp{Seconds: time.Date(2010, 5, 1, 12, 12, 59, 0, time.FixedZone("UTC-8", -8*60*60)).Unix()},
	}

	s.ProtoReflect()

	return map[string]interface{}{
		"student": &s,
		"now":     &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}

}

func TestProtoMessage(t *testing.T) {

	is := is.New(t)
	engine := cel.NewEngine(indigo.CollectDiagnostics(true), indigo.ForceDiagnosticsAllRules(true))
	err := engine.AddRule(makeEducationProtoRules())
	is.NoErr(err)

	results, err := engine.Evaluate(makeStudentProtoData(), "student_actions")
	is.NoErr(err)
	is.Equal(len(results.Results), 3)
	for _, v := range results.Results {
		is.Equal(v.Meta, v.Pass)
	}
}

func TestDiagnosticOptions(t *testing.T) {

	is := is.New(t)

	// Turn off diagnostic collection
	engine := cel.NewEngine(indigo.CollectDiagnostics(false))
	err := engine.AddRule(makeEducationProtoRules())
	is.NoErr(err)

	_, err = engine.Evaluate(makeStudentProtoData(), "student_actions", indigo.ReturnDiagnostics(true))
	if err == nil {
		t.Errorf("Wanted error; should require indigo.CollectDiagnostics to be turned on to enable indigo.ReturnDiagnostics")
	}

	// Turn on diagnostic collection
	engine = cel.NewEngine(indigo.CollectDiagnostics(true))
	err = engine.AddRule(makeEducationProtoRules())
	is.NoErr(err)

	results, err := engine.Evaluate(makeStudentProtoData(), "student_actions", indigo.ReturnDiagnostics(true))
	is.NoErr(err)

	is.Equal(results.RulesEvaluated, 3)

	for _, c := range results.Results {
		fmt.Println(c.Diagnostics)
		is.Equal(c.RulesEvaluated, 1)
		if len(c.Diagnostics) < 100 {
			t.Errorf("Wanted diagnostics for rule %s, got %s", c.RuleID, c.Diagnostics)
		}
	}

}

func TestEvalOptions(t *testing.T) {

	is := is.New(t)

	cases := []struct {
		opts []indigo.EvalOption  // Options to pass to evaluate
		chk  func(*indigo.Result) // Function to check the results
	}{
		{
			opts: []indigo.EvalOption{indigo.MaxDepth(0)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 0) // No child results
			},
		},
		{
			opts: []indigo.EvalOption{indigo.StopIfParentNegative(true)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 0)
			},
		},
		{
			opts: []indigo.EvalOption{indigo.StopIfParentNegative(false)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 4)
			},
		},
		{
			opts: []indigo.EvalOption{indigo.StopFirstPositiveChild(true)},
			chk: func(r *indigo.Result) {
				i := 0
				for _, v := range r.Results {
					if v.Pass {
						i++
					}
				}
				is.Equal(i, 1)
			},
		},
		{
			opts: []indigo.EvalOption{indigo.StopFirstNegativeChild(true)},
			chk: func(r *indigo.Result) {
				i := 0
				for _, v := range r.Results {
					if !v.Pass {
						i++
					}
				}
				is.Equal(i, 1)
			},
		},
		{
			opts: []indigo.EvalOption{indigo.StopFirstNegativeChild(true), indigo.StopFirstPositiveChild(true)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 1)
			},
		},
		{
			opts: []indigo.EvalOption{indigo.ReturnFail(false), indigo.ReturnPass(false)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 0)
			},
		},
		{
			opts: []indigo.EvalOption{indigo.ReturnPass(false)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 2)
			},
		},

		{
			opts: []indigo.EvalOption{indigo.ReturnFail(false)},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 2)
			},
		},
	}

	engine := cel.NewEngine()
	err := engine.AddRule(makeEducationRules()...)
	is.NoErr(err)

	for _, c := range cases {
		result, err := engine.Evaluate(makeStudentData(), "depthRules", c.opts...)
		is.NoErr(err)
		c.chk(result)
	}
}

func TestRuleOptionOverride(t *testing.T) {

	is := is.New(t)

	cases := []struct {
		opts []indigo.EvalOption  // Options to pass to evaluate
		chk  func(*indigo.Result) // Function to check the results
	}{
		{
			opts: []indigo.EvalOption{},
			chk: func(r *indigo.Result) {
				is.Equal(len(r.Results), 3) // rules a, b and E

				// Check how many positives there are in rule D;
				// the rule option should stop at the first positive
				i := 0
				for _, v := range r.Results["D"].Results {
					if v.Pass {
						i++
					}
				}
				is.Equal(i, 1)

				// Check how many positives there are in rule E;
				// without a rule option, all should be returned
				i = 0
				for _, v := range r.Results["E"].Results {
					if v.Pass {
						i++
					}
				}
				is.Equal(i, 2)
			},
		},
	}

	engine := cel.NewEngine()
	err := engine.AddRule(makeEducationRules()...)
	is.NoErr(err)

	for _, c := range cases {
		result, err := engine.Evaluate(makeStudentData(), "ruleOptions", c.opts...)
		is.NoErr(err)
		c.chk(result)
	}
}

func TestRemoveRule(t *testing.T) {
	is := is.New(t)

	engine := cel.NewEngine()
	err := engine.AddRule(makeEducationRules()...)
	is.NoErr(err)

	is.Equal(engine.RuleCount(), 3)

	engine.RemoveRule("student_actions")
	is.Equal(engine.RuleCount(), 2)
}

// ------------------------------------------------------------------------------------------
// BENCHMARKS
//
//
//
//
//

func BenchmarkSimpleRule(b *testing.B) {

	engine := cel.NewEngine()
	education := makeEducationSchema()
	data := makeStudentData()

	rule := indigo.Rule{
		ID:     "student_actions",
		Schema: education,
		Rules: map[string]indigo.Rule{
			"a": {
				ID:     "at_risk",
				Schema: education,
				Expr:   `student.GPA < 2.5 || student.Status == "Probation"`,
			},
		},
	}

	err := engine.AddRule(rule)
	if err != nil {
		b.Errorf("Error adding ruleset: %v", err)
	}

	for i := 0; i < b.N; i++ {
		engine.Evaluate(data, "student_actions")
	}
}

func BenchmarkSimpleRuleWithDiagnostics(b *testing.B) {

	engine := cel.NewEngine(indigo.CollectDiagnostics(true), indigo.ForceDiagnosticsAllRules(true))
	education := makeEducationSchema()
	data := makeStudentData()

	rule := indigo.Rule{
		ID:     "student_actions",
		Schema: education,
		Rules: map[string]indigo.Rule{
			"a": {
				ID:     "at_risk",
				Schema: education,
				Expr:   `student.GPA < 2.5 || student.Status == "Probation"`,
			},
		},
	}

	err := engine.AddRule(rule)
	if err != nil {
		b.Errorf("Error adding ruleset: %v", err)
	}

	for i := 0; i < b.N; i++ {
		engine.Evaluate(data, "student_actions")
	}
}

func BenchmarkRuleWithArray(b *testing.B) {
	engine := cel.NewEngine()
	education := makeEducationSchema()

	rule := indigo.Rule{
		ID:     "student_actions",
		Schema: education,
		Rules: map[string]indigo.Rule{
			"a": {
				ID:     "honors_student",
				Schema: education,
				Expr:   `student.GPA >= 3.6 && student.Status!="Probation" && !("C" in student.Grades)`,
			},
		},
	}

	err := engine.AddRule(rule)
	if err != nil {
		b.Errorf("Error adding ruleset: %v", err)
	}

	data := makeStudentData()
	for i := 0; i < b.N; i++ {
		engine.Evaluate(data, "student_actions")
	}
}

func BenchmarkProtoWithSelf(b *testing.B) {

	pb.DefaultDb.RegisterMessage(&school.Student{})

	schema := indigo.Schema{
		Elements: []indigo.DataElement{
			{Name: "student", Type: indigo.Proto{Protoname: "school.Student", Message: &school.Student{}}},
			{Name: "now", Type: indigo.Timestamp{}},
			{Name: "self", Type: indigo.Proto{Protoname: "school.HonorsConfiguration", Message: &school.HonorsConfiguration{}}},
		},
	}

	engine := cel.NewEngine()

	rule := indigo.Rule{
		ID:     "student_actions",
		Schema: schema,
		Rules: map[string]indigo.Rule{
			"a": {
				ID:   "at_risk",
				Expr: `student.GPA < self.Minimum_GPA || student.Status == school.Student.status_type.PROBATION`,
				Self: &school.HonorsConfiguration{Minimum_GPA: 3.7},
				Meta: false,
			},
		},
	}

	err := engine.AddRule(rule)
	if err != nil {
		log.Fatalf("Error adding ruleset: %v", err)
	}

	s := school.Student{
		Age:            16,
		GPA:            3.76,
		Status:         school.Student_ENROLLED,
		Grades:         []float64{4.0, 4.0, 3.7},
		Attrs:          map[string]string{"Nickname": "Joey"},
		EnrollmentDate: &timestamp.Timestamp{Seconds: time.Date(2010, 5, 1, 12, 12, 59, 0, time.FixedZone("UTC-8", -8*60*60)).Unix()},
	}

	data := map[string]interface{}{
		"student": &s,
		"now":     &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}

	for i := 0; i < b.N; i++ {
		engine.Evaluate(data, "student_actions")
	}

}

func BenchmarkProtoWithoutSelf(b *testing.B) {

	pb.DefaultDb.RegisterMessage(&school.Student{})

	schema := indigo.Schema{
		Elements: []indigo.DataElement{
			{Name: "student", Type: indigo.Proto{Protoname: "school.Student", Message: &school.Student{}}},
			{Name: "now", Type: indigo.Timestamp{}},
		},
	}

	engine := cel.NewEngine()

	rule := indigo.Rule{
		ID:     "student_actions",
		Schema: schema,
		Rules: map[string]indigo.Rule{
			"a": {
				ID:   "at_risk",
				Expr: `student.GPA < 2.5 || student.Status == school.Student.status_type.PROBATION`,
				Meta: false,
			},
		},
	}

	err := engine.AddRule(rule)
	if err != nil {
		log.Fatalf("Error adding ruleset: %v", err)
	}

	s := school.Student{
		Age:            16,
		GPA:            3.76,
		Status:         school.Student_ENROLLED,
		Grades:         []float64{4.0, 4.0, 3.7},
		Attrs:          map[string]string{"Nickname": "Joey"},
		EnrollmentDate: &timestamp.Timestamp{Seconds: time.Date(2010, 5, 1, 12, 12, 59, 0, time.FixedZone("UTC-8", -8*60*60)).Unix()},
	}

	data := map[string]interface{}{
		"student": &s,
		"now":     &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}

	for i := 0; i < b.N; i++ {
		engine.Evaluate(data, "student_actions")
	}

}

func BenchmarkCalculation(b *testing.B) {
	engine := cel.NewEngine()

	education := makeEducationSchema()
	data := makeStudentData()

	for i := 0; i < b.N; i++ {
		_, err := engine.Calculate(data, `2.0+student.GPA + (1.344 * student.Adjustment)/3.3`, education)
		if err != nil {
			b.Fatalf("Could not calculate risk factor: %v", err)
		}
	}
}
