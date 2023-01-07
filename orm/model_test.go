package orm

import (
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseModel(t *testing.T) {
	type TestModel struct {
		Id        int64
		FirstName string
		Age       int8
		LastName  *sql.NullString
	}

	testCases := []struct {
		name      string
		val       any
		wantModel *model
		wantErr   error
	}{
		{
			name:    "test model",
			val:     TestModel{},
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			// 指针
			name: "pointer",
			val:  &TestModel{},
			wantModel: &model{
				tableName: "test_model",
				fieldMap: map[string]*field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"Age": {
						colName: "age",
					},
					"LastName": {
						colName: "last_name",
					},
				},
			},
		},
		{
			// 多级指针
			name: "multiple pointer",
			// 因为 Go 编译器的原因，所以我们写成这样
			val: func() any {
				val := &TestModel{}
				return &val
			}(),
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:    "map",
			val:     map[string]string{},
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:    "slice",
			val:     []int{},
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
		{
			name:    "basic type",
			val:     0,
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 *User"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := parseModel(tc.val)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func Test_underscoreName(t *testing.T) {
	testCases := []struct {
		name    string
		srcStr  string
		wantStr string
	}{
		{
			name:    "upper cases",
			srcStr:  "ID",
			wantStr: "i_d",
		},
		{
			name:    "use number",
			srcStr:  "Table1Name",
			wantStr: "table1_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := underscoreName(tc.srcStr)
			assert.Equal(t, tc.wantStr, res)
		})
	}
}
