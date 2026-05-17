package notifications

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_NewTask(t *testing.T) {
	msg := "**🎓 Новые задачи**\n\n\n[Итоговая оценка (Инструменты разработчика)]" +
		"(https://my.centraluniversity.ru/learn/courses/view/actual/527/themes/7042/longreads/13244)"

	n := Parse("p1", msg, time.Now())
	require.NotNil(t, n)
	assert.Equal(t, KindNewTask, n.Kind)
	assert.Equal(t, 527, n.CourseID)
	assert.Equal(t, 7042, n.ThemeID)
	assert.Equal(t, 13244, n.LongreadID)
	assert.Contains(t, n.Title, "Итоговая оценка")
}

func TestParse_Graded(t *testing.T) {
	msg := "**🎓 Задача оценена**\nЗадача оценена на 10 баллов\n\n" +
		"[ДЗ 5 (Алгоритмы. Часть 1.)](https://my.centraluniversity.ru/learn/courses/view/actual/528/themes/6284/longreads/12029)"

	n := Parse("p2", msg, time.Now())
	require.NotNil(t, n)
	assert.Equal(t, KindGraded, n.Kind)
	assert.Equal(t, 528, n.CourseID)
	assert.Equal(t, 12029, n.LongreadID)
}

func TestParse_Skipped(t *testing.T) {
	assert.Nil(t, Parse("p3", "**🎓 Открылась запись на пары**\nfoo", time.Now()))
}
