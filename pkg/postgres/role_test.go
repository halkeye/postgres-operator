package postgres

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestIsPQError(t *testing.T) {
	genericError := errors.New("regular error")

	pqError := new(pq.Error)
	pqError.Code = "42710"

	if isPQError(genericError, "42710") {
		t.Fatalf("generic error is not an pq error")
	}

	if isPQError(pqError, "111111") {
		t.Fatalf("might be a pq error, but isn't this specific one")
	}

	if !isPQError(pqError, "42710") {
		t.Fatalf("should match targeted pqerror")
	}

	// multiple
	if !isPQError(pqError, "42710", "111111") {
		t.Fatalf("should match targeted pqerror")
	}
}

func TestCreateGroupRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	duplicateObjectPGError := new(pq.Error)
	duplicateObjectPGError.Code = "42710" // duplicate_object

	diskFullError := new(pq.Error)
	diskFullError.Code = "53100" // disk_full

	mock.ExpectExec(`CREATE ROLE "working"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`CREATE ROLE "already-exists"`).WillReturnError(duplicateObjectPGError)
	mock.ExpectExec(`CREATE ROLE "disk-full"`).WillReturnError(diskFullError)

	pgObj := &pg{db: db}
	if err := pgObj.CreateGroupRole("working"); err != nil {
		t.Errorf("error calling create role: %s", err)
	}

	if err := pgObj.CreateGroupRole("already-exists"); err != nil {
		t.Errorf("error calling create role: %s", err)
	}

	if err := pgObj.CreateGroupRole("disk-full"); err == nil {
		t.Errorf("non ignored error didn't throw error")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
