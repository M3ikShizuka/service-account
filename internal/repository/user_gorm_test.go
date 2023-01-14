package repository

import (
	"context"
	"encoding/hex"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	logger2 "gorm.io/gorm/logger"
	"regexp"
	"service-account/internal/domain"
	"testing"
	"time"
)

type TestTable struct {
	name           string
	args           TestTableArgs
	mock           TestTableMock
	expectedResult TestTableExpectedResult
}

type TestTableArgs struct {
	user *domain.User
}

type TestTableMock struct {
	expectedQuery func(mock sqlmock.Sqlmock, tt *TestTable)
}

type TestTableExpectedResult struct {
	err    error
	usedId uint32
}

func TestUser_Create(t *testing.T) {
	testPassHash, _ := hex.DecodeString("92b2723f184a5f9b17ba52b88079391b") // pass: 1234567890qwerty
	const sqlRequest = `INSERT INTO "tb_users"`

	tests := []TestTable{
		{
			name: "Create user",
			args: TestTableArgs{
				user: &domain.User{
					Id:               1,
					Username:         "test",
					Email:            "test@mail.com",
					PasswordHash:     testPassHash,
					DateRegistration: time.Now(),
					DateLastOnline:   time.Now(),
				},
			},
			mock: TestTableMock{
				expectedQuery: func(mock sqlmock.Sqlmock, tt *TestTable) {
					mock.ExpectBegin()

					//// Change mock.ExpectExec to mock.ExpectQuery.
					//// SRC: https://stackoverflow.com/a/60982925
					//// Issue: https://github.com/DATA-DOG/go-sqlmock/issues/118
					//// Article: https://betterprogramming.pub/how-to-unit-test-a-gorm-application-with-sqlmock-97ee73e36526
					mock.ExpectQuery(
						regexp.QuoteMeta(sqlRequest)).
						WithArgs(
							tt.args.user.Username,
							tt.args.user.Email,
							tt.args.user.PasswordHash,
							tt.args.user.DateRegistration,
							tt.args.user.DateLastOnline,
							tt.args.user.Id,
						).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).
							AddRow(tt.args.user.Id))

					mock.ExpectCommit()
				},
			},
			expectedResult: TestTableExpectedResult{
				err:    nil,
				usedId: 1,
			},
		},
		{
			name: "User already exist",
			args: TestTableArgs{
				user: &domain.User{
					Id:               1,
					Username:         "test",
					Email:            "test@mail.com",
					PasswordHash:     testPassHash,
					DateRegistration: time.Now(),
					DateLastOnline:   time.Now(),
				},
			},
			mock: TestTableMock{
				expectedQuery: func(mock sqlmock.Sqlmock, tt *TestTable) {
					mock.ExpectBegin()

					//// Change mock.ExpectExec to mock.ExpectQuery.
					//// SRC: https://stackoverflow.com/a/60982925
					//// Issue: https://github.com/DATA-DOG/go-sqlmock/issues/118
					//// Article: https://betterprogramming.pub/how-to-unit-test-a-gorm-application-with-sqlmock-97ee73e36526
					mock.ExpectQuery(
						regexp.QuoteMeta(sqlRequest)).
						WithArgs(
							tt.args.user.Username,
							tt.args.user.Email,
							tt.args.user.PasswordHash,
							tt.args.user.DateRegistration,
							tt.args.user.DateLastOnline,
							tt.args.user.Id,
						).
						WillReturnError(ErrRecordAlreadyExist)

					mock.ExpectRollback()
				},
			},
			expectedResult: TestTableExpectedResult{
				err:    ErrRecordAlreadyExist,
				usedId: 0,
			},
		},
	}

	// Init mockDB mock.
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		return
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open(
		postgres.New(
			postgres.Config{
				Conn: mockDB,
			}),
		&gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
		return
	}

	gormDB.Logger.LogMode(logger2.Info)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Expected behavior.
			tt.mock.expectedQuery(mock, &tt)

			// Call test function.
			ctx := context.Background()
			r := UserRepositoryGorm{
				db: gormDB,
			}

			err = r.Create(ctx, tt.args.user)
			if err != nil {
				assert.Equal(t, tt.expectedResult.err, err)
			} else {
				assert.Equal(t, tt.args.user.Id, tt.expectedResult.usedId)
			}

			// We make sure that all expectations were met.
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
