package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		fullName      string
		stars         int
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid repository",
			fullName:    "owner/repo",
			stars:       100,
			expectError: false,
		},
		{
			name:          "Invalid format - no slash",
			fullName:      "owner-repo",
			stars:         100,
			expectError:   true,
			expectedError: "invalid repository full name format: owner-repo",
		},
		{
			name:          "Invalid format - missing owner",
			fullName:      "/repo",
			stars:         100,
			expectError:   true,
			expectedError: "invalid repository full name format: /repo",
		},
		{
			name:          "Invalid format - missing repo",
			fullName:      "owner/",
			stars:         100,
			expectError:   true,
			expectedError: "invalid repository full name format: owner/",
		},
		{
			name:          "Invalid stars - negative",
			fullName:      "owner/repo",
			stars:         -1,
			expectError:   true,
			expectedError: "stars count cannot be negative: -1",
		},
		{
			name:        "Valid stars - zero",
			fullName:    "owner/repo",
			stars:       0,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, err := NewRepository(tc.fullName, tc.stars)

			if tc.expectError {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
				assert.Nil(t, repo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, repo)
				assert.Equal(t, tc.fullName, repo.FullName)
				assert.Equal(t, tc.stars, repo.Stars)
			}
		})
	}
}
