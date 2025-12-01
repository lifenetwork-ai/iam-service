package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePhoneE164_require(t *testing.T) {
	type tc struct {
		name          string
		in            string
		defaultRegion string
		wantE164      string
		wantRegion    string
		wantErr       error
	}

	tests := []tc{
		// --- Vietnam ---
		{
			name:          "VN redundant 0 after country code -> E164",
			in:            "+840344381024",
			defaultRegion: "VN",
			wantE164:      "+84344381024",
			wantRegion:    "VN",
		},
		{
			name:          "VN national mobile -> E164",
			in:            "0344381024",
			defaultRegion: "VN",
			wantE164:      "+84344381024",
			wantRegion:    "VN",
		},
		{
			name:          "VN with spaces/dashes -> E164",
			in:            "0 344-381-024",
			defaultRegion: "VN",
			wantE164:      "+84344381024",
			wantRegion:    "VN",
		},
		{
			name:          "VN with 00 prefix -> E164",
			in:            "00 84 344 381 024",
			defaultRegion: "VN",
			wantE164:      "+84344381024",
			wantRegion:    "VN",
		},

		// --- Thailand ---
		{
			name:          "TH in E164 with spaces",
			in:            "+66 81 234 5678",
			defaultRegion: "VN", // has '+', defaultRegion irrelevant
			wantE164:      "+66812345678",
			wantRegion:    "TH",
		},

		// --- Indonesia ---
		{
			name:          "ID national without '+' using defaultRegion ID",
			in:            "081234567890",
			defaultRegion: "ID",
			wantE164:      "+6281234567890",
			wantRegion:    "ID",
		},
		{
			name:          "ID national without '+' but wrong defaultRegion -> invalid",
			in:            "081234567890",
			defaultRegion: "VN",
			wantErr:       errInvalidPhone,
		},

		// --- Korea ---
		{
			name:          "KR national mobile -> E164",
			in:            "010-1234-5678",
			defaultRegion: "KR",
			wantE164:      "+821012345678",
			wantRegion:    "KR",
		},

		// --- China ---
		{
			name:          "CN national mobile -> E164",
			in:            "13800138000",
			defaultRegion: "CN",
			wantE164:      "+8613800138000",
			wantRegion:    "CN",
		},

		// --- US (now allowed because no allowlist) ---
		{
			name:          "US number with '+' -> allowed",
			in:            "+14155552671",
			defaultRegion: "VN",
			wantE164:      "+14155552671",
			wantRegion:    "US",
		},

		// --- Invalid / Global service ---
		{
			name:          "Invalid country code +80",
			in:            "+801234",
			defaultRegion: "VN",
			wantErr:       errInvalidPhone,
		},
		{
			name:          "Global service +800 UIFN -> reject",
			in:            "+80012345678",
			defaultRegion: "VN",
			wantErr:       errInvalidPhone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotE164, gotRegion, err := NormalizePhoneE164(tt.in, tt.defaultRegion)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantE164, gotE164)
			require.Equal(t, tt.wantRegion, gotRegion)
		})
	}
}

func TestIsPhoneE164_require(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"+84344381024", true},
		{"+66812345678", true},
		{"+6281234567890", true},
		{"+821012345678", true},
		{"+8613800138000", true},
		{"0344381024", false},
		{"+80", false},
		{"  +84344381024  ", true},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			require.Equal(t, c.want, IsPhoneE164(c.in))
		})
	}
}
