package client

type PingFederateUser struct {
	Email       string   `json:"emailAddress,omitempty"`
	EncryptedPassword string   `json:"encryptedPassword"`
	Username    string   `json:"username"`
	PhoneNumber string   `json:"phoneNumber,omitempty"`
	Department  string   `json:"department,omitempty"`
	Description string   `json:"description"`
	IsAuditor   bool     `json:"auditor"`
	IsActive    bool     `json:"active"`
	Roles       []string `json:"roles"`
}

type getAdminUsersResponse struct {
	Items []PingFederateUser `json:"items"`
}

type PingFederateRole struct {
	Name        string   `json:"name"`
	ID          string   `json:"id"`
}
/*
{
	"items": [
	  {
		"username": "Administrator",
		"encryptedPassword": "OBF:JWE:eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2Iiwia2lkIjoiR3gySkc3anBBRSIsInZlcnNpb24iOiIxMi4xLjEuMCJ9..kWHdfXGNGzyIk_WdxsSrCw.2imDyNbsjFgxzPkUjz07eQihzc4DJd6gzrFXGwvUajPHorWelfIzrGEyGMxazmGOkeDyYOHz30j8TJb33xWvAg.SExW5L5YUs-aT2W_W7zovQ",
		"description": "Initial administrator user",
		"auditor": false,
		"active": true,
		"roles": [
		  "USER_ADMINISTRATOR",
		  "EXPRESSION_ADMINISTRATOR",
		  "ADMINISTRATOR",
		  "CRYPTO_ADMINISTRATOR"
		]
	  },
	  {
		"username": "sam-ng",
		"encryptedPassword": "OBF:JWE:eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2Iiwia2lkIjoiR3gySkc3anBBRSIsInZlcnNpb24iOiIxMi4xLjEuMCJ9..wT59Q5juJfHcoVW-Ix7UGQ.5NF0cx1O-Rwoq80-oCt6l1Es8N47nDtyTIhzHKncP010gJQ4QhVRApAKu3FNFX1vHP-W0Ffc9Zw0ZnOu3tMCvA.o_K-OvJcT9j_jzJK8Bjh_Q",
		"phoneNumber": "1234567890",
		"emailAddress": "sam@nyedis.com",
		"department": "iam",
		"description": "test admin user 1",
		"auditor": false,
		"active": true,
		"roles": [
		  "USER_ADMINISTRATOR",
		  "EXPRESSION_ADMINISTRATOR",
		  "ADMINISTRATOR",
		  "CRYPTO_ADMINISTRATOR"
		]
	  },
	  {
		"username": "kurt-bitner",
		"encryptedPassword": "OBF:JWE:eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2Iiwia2lkIjoiR3gySkc3anBBRSIsInZlcnNpb24iOiIxMi4xLjEuMCJ9..aKVFJDuLvBeE44N67HvfiA.B10jH7kzhaOb--yd_Xe-ntRBvWdEOzoUk2ME8jnHQ7Hy4nzjz2xetgyCFXuxAz8X8O7bhSEsaoLG0US9P0XTHw.vk6qn9BX9PPHwWWPgJ9QNA",
		"phoneNumber": "1234567891",
		"emailAddress": "kurt@nyedis.com",
		"department": "iam",
		"description": "test admin user 2",
		"auditor": true,
		"active": true,
		"roles": []
	  }
	]
  }
*/