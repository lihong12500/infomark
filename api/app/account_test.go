// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package app

import (
	"github.com/cgtuebingen/infomark-backend/api/helper"
	authpkg "github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/database"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func TestAccount(t *testing.T) {

	email.DefaultMail = email.VoidMail

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	userStore := database.NewUserStore(db)
	rs := NewAccountResource(userStore)

	g.Describe("GetAccount", func() {
		g.It("Should require valid claims", func() {

			re := helper.Payload{
				Data:   helper.H{},
				Method: "POST",
			}

			w := helper.SimulateRequest(re, rs.GetHandler, authenticate.RequiredValidAccessClaims)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should return info when claims are invalid", func() {

			re := helper.Payload{
				Data:         helper.H{},
				Method:       "POST",
				AccessClaims: authenticate.NewAccessClaims(0, true), // 0 is invalid
			}

			w := helper.SimulateRequest(re, rs.GetHandler, authenticate.RequiredValidAccessClaims)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should return info when claims are valid", func() {

			re := helper.Payload{
				Data:         helper.H{},
				Method:       "POST",
				AccessClaims: authenticate.NewAccessClaims(1, true),
			}

			w := helper.SimulateRequest(re, rs.GetHandler, authenticate.RequiredValidAccessClaims)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

	})

}

func TestAccountCreate(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	userStore := database.NewUserStore(db)
	rs := NewAccountResource(userStore)

	g.Describe("Account Creation", func() {
		g.It("Should not create invalid accounts", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{
							"email":          "foo@test.com",
							"plain_password": "bar",
						},
						"user": helper.H{
							"first_name": "",
						},
					},
					Method: "POST",
				},
				rs.CreateHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should not create accounts with too short password", func() {

			min_len := viper.GetInt("min_password_length")
			too_short_password := authpkg.GenerateToken(min_len - 1)

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{
							"email":          "foo@test.com",
							"plain_password": too_short_password,
						},
						"user": helper.H{
							"first_name": "",
						},
					},
					Method: "POST",
				},
				rs.CreateHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid accounts", func() {

			min_len := viper.GetInt("min_password_length")
			ok_password := authpkg.GenerateToken(min_len)

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{
							"email":          "foo@test.com",
							"plain_password": ok_password,
						},
						"user": helper.H{
							"first_name":     "Max",
							"last_name":      "Mustermensch",
							"semester":       2,
							"student_number": "0815",
							"subject":        "math",
							"email":          "foo@test.com",
						},
					},
					Method: "POST",
				},
				rs.CreateHandler,
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			user_after, err := userStore.FindByEmail("foo@test.com")
			g.Assert(err).Equal(nil)
			g.Assert(user_after.Email).Equal("foo@test.com")
			g.Assert(user_after.ConfirmEmailToken.Valid).Equal(true)
			g.Assert(user_after.FirstName).Equal("Max")
			g.Assert(user_after.LastName).Equal("Mustermensch")
			g.Assert(user_after.Semester).Equal(2)
			g.Assert(user_after.StudentNumber).Equal("0815")
			g.Assert(user_after.Subject).Equal("math")

			password_valid := authpkg.CheckPasswordHash(ok_password, user_after.EncryptedPassword)
			g.Assert(password_valid).Equal(true)

		})
	})
}

func TestAccountChanges(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	userStore := database.NewUserStore(db)
	rs := NewAccountResource(userStore)

	g.Describe("Account Changes", func() {
		g.It("Should require valid access claims", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"first_name": "foo",
						"last_name":  "bar",
					},
					Method: "PATCH",
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should not change the account when incorrect password", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{},
						"user": helper.H{
							"email": "foo",
						},
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should change email with correct password and confirm token", func() {

			user_before, err := userStore.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_before.Email).Equal("test@uni-tuebingen.de")

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{
							"plain_password": "test",
						},
						"user": helper.H{
							"email":      "foo@uni-tuebingen.de",
							"first_name": "Peter",
							"last_name":  "Zwegat",
						},
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusNoContent)

			user_after, err := userStore.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_after.Email).Equal("foo@uni-tuebingen.de")
			g.Assert(user_after.FirstName).Equal("Peter")
			g.Assert(user_after.LastName).Equal("Zwegat")
			g.Assert(user_after.ConfirmEmailToken.Valid).Equal(true)
		})
	})
}
