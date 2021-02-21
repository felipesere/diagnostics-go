package diagnostics

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCanStartOutAsNothing(t *testing.T) {
	err := None()
	require.False(t, err.IsErr())
	require.Nil(t, err.Err())

	err = FromErr(nil)
	require.False(t, err.IsErr())
	require.Nil(t, err.Err())
}

func TestCanBeCreated(t *testing.T) {
	err := FromString("this failed")

	require.True(t, err.IsErr())
	require.NotNil(t, err.Err())
	require.Equal(t, `this failed`, err.UserFacing())
}

func TestCanWrapAnInnerError(t *testing.T) {
	err := FromErr(errors.New("some inner error"))

	output := err.UserFacing()

	require.Equal(t, `some inner error`, output)
}

func TestCanWrap(t *testing.T) {
	err := FromErr(errors.New("some inner error")).Wrap("trying to do this")

	output := err.UserFacing()

	require.Equal(t, `trying to do this
    ┗━ some inner error`, output)

	moreErr := err.Wrap("and some more context")
	output = moreErr.UserFacing()
	require.Equal(t, `and some more context
    ┗━ trying to do this
        ┗━ some inner error`, output)
}

func TestCanHoldSomeData(t *testing.T) {
	err := FromErr(errors.New("some inner error")).WithData("foo", "bar").WithData("bar", 12)

	output := err.UserFacing()

	require.Equal(t, `some inner error: bar = 12, foo = "bar"`, output)
}

func TestCanHoldBunchOfData(t *testing.T) {
	err := FromErr(errors.New("some inner error")).WithAllData(map[string]interface{}{
		"foo": "bar",
		"bar": 12,
	})

	output := err.UserFacing()

	require.Equal(t, `some inner error: bar = 12, foo = "bar"`, output)

	moreData := err.WithAllData(map[string]interface{}{
		"bar":  42,
		"batz": false,
	})
	require.Equal(t, `some inner error: bar = 42, batz = false, foo = "bar"`, moreData.UserFacing())
}

func TestDifferentPartsHoldData(t *testing.T) {
	err := FromString("some inner error").WithData("foo", 12).Wrap("outer error").WithData("bar", true)

	output := err.UserFacing()

	require.Equal(t, `outer error: bar = true
    ┗━ some inner error: foo = 12`, output)
}

func TestIsStillAnError(t *testing.T) {
	origin := errors.New("the origin")

	wrapped := FromErr(origin).Wrap("layer").Wrap("deeper").Wrap("Very top")

	require.True(t, errors.Is(wrapped, origin))
	require.False(t, errors.Is(wrapped, errors.New("random")))
}

type SampleErr struct {
	age int
}

func (s SampleErr) Error() string {
	return fmt.Sprintf("The age is %d", s.age)
}

func TestExtractsDataAsError(t *testing.T) {
	origin := SampleErr{age: 42}

	wrapped := FromErr(origin).Wrap("layer").Wrap("deeper").Wrap("Very top")

	var extracted SampleErr
	require.True(t, errors.As(wrapped, &extracted))
	require.Equal(t, 42, extracted.age)

	other := FromString("oh no!").Wrap("sad!")
	require.False(t, errors.As(other, &extracted))
}

func TestLossyConversionToErrorString(t *testing.T) {
	err := FromString("some inner error").WithData("foo", 12)
	require.Equal(t, `some inner error`, err.Error())

	err = err.Wrap("outer error").WithData("bar", true)
	require.Equal(t, `outer error: some inner error`, err.Error())

	err = FromErr(err).Wrap("last one")
	require.Equal(t, `last one: outer error: some inner error`, err.Error())
}
