package recordings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_SingleItem(t *testing.T) {
	msg := "**🎓 Записи занятий**\nПривет!\nНовые записи твоих занятий за 23.12.2025:\n\n" +
		"1. Пара: Инструменты разработчика\n" +
		"   Время: 19:00-21:55\n" +
		"   Ссылки:\n" +
		"   - https://centraluniversity.ktalk.ru/recordings/PaWLmfACTawEJ5owQu63 (размер: 83,3 МБ)\n   \n\n"

	posted := time.Date(2025, 12, 23, 15, 0, 0, 0, time.UTC)
	got := Parse(msg, posted)

	require.Len(t, got, 1)
	r := got[0]
	assert.Equal(t, "Инструменты разработчика", r.Subject)
	assert.Equal(t, "19:00", r.StartTime)
	assert.Equal(t, "21:55", r.EndTime)
	assert.Equal(t, 2025, r.Date.Year())
	assert.Equal(t, time.December, r.Date.Month())
	assert.Equal(t, 23, r.Date.Day())
	require.Len(t, r.Links, 1)
	assert.Equal(t, "https://centraluniversity.ktalk.ru/recordings/PaWLmfACTawEJ5owQu63", r.Links[0].URL)
	assert.Equal(t, "83,3 МБ", r.Links[0].Size)
}

func TestParse_MultipleItems(t *testing.T) {
	msg := "**🎓 Записи занятий**\nНовые записи твоих занятий за 1.2.2026:\n\n" +
		"1. Пара: Разработка на языке программирования Go\n" +
		"   Время: 13:00-14:20\n" +
		"   Ссылки:\n" +
		"   - https://centraluniversity.ktalk.ru/recordings/7kvIkpKDV0e1E8Klc5FL (размер: 24,9 МБ)\n\n" +
		"2. Пара: Алгоритмы\n" +
		"   Время: 15:00-16:20\n" +
		"   Ссылки:\n" +
		"   - https://centraluniversity.ktalk.ru/recordings/AAA\n"

	got := Parse(msg, time.Now())
	require.Len(t, got, 2)
	assert.Equal(t, "Разработка на языке программирования Go", got[0].Subject)
	assert.Equal(t, "Алгоритмы", got[1].Subject)
	assert.Empty(t, got[1].Links[0].Size)
}

func TestParse_NonRecordingPost(t *testing.T) {
	got := Parse("**🎓 Новые задачи**\nfoo", time.Now())
	assert.Nil(t, got)
}
