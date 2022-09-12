package store

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"testing"
	"webserver/clock"
	"webserver/entity"
	"webserver/testutil"
)

func prepareTasks(ctx context.Context, t *testing.T, con Execer) entity.Tasks {
	t.Helper()
	// 一度きれいにしておく
	if _, err := con.ExecContext(ctx, "DELETE FROM task;"); err != nil {
		t.Logf("failed to initialize task: %v", err)
	}
	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{
			Title: "want task 1", Status: "todo",
			CreatedAt: c.Now(), UpdatedAt: c.Now(),
		},
		{
			Title: "want task 2", Status: "todo",
			CreatedAt: c.Now(), UpdatedAt: c.Now(),
		},
		{
			Title: "want task 3", Status: "done",
			CreatedAt: c.Now(), UpdatedAt: c.Now(),
		},
	}
	result, err := con.ExecContext(ctx,
		`INSERT INTO task (title, status, created_at, updated_at)
			VALUES
			    (?, ?, ?, ?),
			    (?, ?, ?, ?),
			    (?, ?, ?, ?);`,
		wants[0].Title, wants[0].Status, wants[0].CreatedAt, wants[0].UpdatedAt,
		wants[1].Title, wants[1].Status, wants[1].CreatedAt, wants[1].UpdatedAt,
		wants[2].Title, wants[2].Status, wants[2].CreatedAt, wants[2].UpdatedAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	wants[0].ID = entity.TaskID(id)
	wants[1].ID = entity.TaskID(id + 1)
	wants[2].ID = entity.TaskID(id + 2)
	return wants
}

func TestRepository_ListTasks(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.OpenDBForTest(t).BeginTxx(ctx, nil)
	t.Cleanup(func() {
		_ = tx.Rollback()
	})
	if err != nil {
		return
	}

	wants := prepareTasks(ctx, t, tx)
	sut := &Repository{}
	gots, err := sut.ListTasks(ctx, tx)
	if err != nil {
		return
	}
	if err != nil {
		t.Fatalf("unexected error: %v", err)
	}
	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s", d)
	}
}

func TestRepository_AddTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c := clock.FixedClocker{}
	var wantID int64 = 20
	okTask := &entity.Task{
		Title:     "ok task",
		Status:    "todo",
		CreatedAt: c.Now(),
		UpdatedAt: c.Now(),
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectExec(
		// エスケープが必要
		`INSERT INTO task \(title, status, created_at, updated_at\) VALUES \(\?, \?, \?, \?\)`,
	).WithArgs(okTask.Title, okTask.Status, c.Now(), c.Now()).
		WillReturnResult(sqlmock.NewResult(wantID, 1))

	xdb := sqlx.NewDb(db, "mysql")
	r := &Repository{Clocker: c}
	if err := r.AddTask(ctx, xdb, okTask); err != nil {
		t.Errorf("want no error, but got %v", err)
	}
}
