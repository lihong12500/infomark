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
  "encoding/json"
  "net/http"
  "testing"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
)

func TestGroup(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Group", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {

      w := tape.Get("/api/v1/courses/1/groups")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/groups", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all groups from a course", func() {
      w := tape.GetWithClaims("/api/v1/courses/1/groups", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      groups_actual := []model.Group{}
      err := json.NewDecoder(w.Body).Decode(&groups_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(groups_actual)).Equal(10)
    })

    g.It("Should get a specific group", func() {
      entry_expected, err := stores.Group.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/groups/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_actual := &model.Group{}
      err = json.NewDecoder(w.Body).Decode(entry_actual)
      g.Assert(err).Equal(nil)

      g.Assert(entry_actual.ID).Equal(entry_expected.ID)
      g.Assert(entry_actual.TutorID).Equal(entry_expected.TutorID)
      g.Assert(entry_actual.CourseID).Equal(entry_expected.CourseID)
      g.Assert(entry_actual.Description).Equal(entry_expected.Description)
    })

    g.It("Creating should require claims", func() {
      w := tape.Post("/api/v1/courses/1/groups", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.Xit("Creating should require body", func() {
      // TODO empty request with claims
    })

    g.It("Should create valid group", func() {
      entries_before, err := stores.Group.GroupsOfCourse(1)
      g.Assert(err).Equal(nil)

      entry_sent := model.Group{
        TutorID:     1,
        CourseID:    1,
        Description: "blah blahe",
      }

      err = entry_sent.Validate()
      g.Assert(err).Equal(nil)

      w := tape.PostWithClaims("/api/v1/courses/1/groups", helper.ToH(entry_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      entry_return := &model.Group{}
      err = json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(entry_return.TutorID).Equal(entry_sent.TutorID)
      g.Assert(entry_return.CourseID).Equal(entry_sent.CourseID)
      g.Assert(entry_return.Description).Equal(entry_sent.Description)

      entries_after, err := stores.Group.GroupsOfCourse(1)
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) + 1)
    })

    g.It("Should update a group", func() {
      // group (id=1) belongs to course(id=1)
      entry_sent := model.Group{
        TutorID:     9,
        CourseID:    1,
        Description: "new descr",
      }

      // students
      w := tape.PlayDataWithClaims("PUT", "/api/v1/groups/1", tape.ToH(entry_sent), 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PlayDataWithClaims("PUT", "/api/v1/groups/1", tape.ToH(entry_sent), 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PlayDataWithClaims("PUT", "/api/v1/groups/1", tape.ToH(entry_sent), 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Group.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.TutorID).Equal(entry_sent.TutorID)
      g.Assert(entry_after.CourseID).Equal(entry_sent.CourseID)
      g.Assert(entry_after.CourseID).Equal(entry_sent.CourseID)
    })

    g.It("Should delete when valid access claims", func() {
      entries_before, err := stores.Group.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/groups/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // verify nothing has changes
      entries_after, err := stores.Group.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      // students
      w = tape.DeleteWithClaims("/api/v1/groups/1", 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.DeleteWithClaims("/api/v1/groups/1", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.DeleteWithClaims("/api/v1/groups/1", 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a sheet less exists
      entries_after, err = stores.Group.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)
    })

    g.It("Find my group when being a student", func() {
      // a random student (checked via pgweb)
      loginID := int64(112)

      w := tape.Get("/api/v1/courses/1/group")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/group", loginID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_return := &model.Group{}
      err := json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(err).Equal(nil)

      // we cannot check the other entries
      g.Assert(entry_return.CourseID).Equal(int64(1))
    })

    g.It("Find my group when being a tutor", func() {
      // a random student (checked via pgweb)
      loginID := int64(2)

      w := tape.Get("/api/v1/courses/1/group")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/group", loginID, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_return := &model.Group{}
      err := json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(err).Equal(nil)

      // we cannot check the other entries
      g.Assert(entry_return.CourseID).Equal(int64(1))
      g.Assert(entry_return.TutorID).Equal(loginID)
    })

    g.It("Permission test", func() {
      url := "/api/v1/courses/1/groups"

      // global root can do whatever they want
      w := tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // enrolled tutors can access
      w = tape.GetWithClaims(url, 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // enrolled students can access
      w = tape.GetWithClaims(url, 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // disenroll student
      w = tape.DeleteWithClaims("/api/v1/courses/1/enrollments", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // cannot access anymore
      w = tape.GetWithClaims(url, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })

  })

}
