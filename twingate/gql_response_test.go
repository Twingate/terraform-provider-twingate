package twingate

// func TestParseErrors(t *testing.T) {
// 	t.Run("Test Twingate Resource : Parse Errors", func(t *testing.T) {

// 		msg0 := graphql.String("test response 0")
// 		msg1 := graphql.String("test response 1")
// 		f := []graphql.String{msg0, msg1}
// 		Locations := []*queryErrorsLocation{}
// 		Location := &queryErrorsLocation{
// 			Line:   1,
// 			Column: 2,
// 		}
// 		Locations = append(Locations, Location)
// 		Errors := []*queryErrors{}
// 		var Path []graphql.String
// 		Error0 := &queryErrors{
// 			Message:   graphql.String(msg0),
// 			Locations: Locations,
// 			Path:      Path,
// 		}
// 		Error1 := &queryErrors{
// 			Message:   graphql.String(msg1),
// 			Locations: Locations,
// 			Path:      Path,
// 		}
// 		Errors = append(Errors, Error0)
// 		Errors = append(Errors, Error1)

// 		messages := parseErrors(Errors)

// 		assert.Equal(t, f, messages)
// 	})
// }
