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
  "time"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
)

func countEnrollments(tape *Tape, stmt string, courseID int64) (int, error) {
  var number_enrollments_expected int
  err := tape.DB.Get(&number_enrollments_expected, stmt, courseID)
  return number_enrollments_expected, err
}

func TestCourse(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Course", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require claims", func() {

      w := tape.Get("/api/v1/courses")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

    })

    g.It("Should list all courses", func() {
      w := tape.GetWithClaims("/api/v1/courses", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      courses_actual := []model.Course{}
      err := json.NewDecoder(w.Body).Decode(&courses_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(courses_actual)).Equal(2)
    })

    g.It("Should get a specific course", func() {

      w := tape.GetWithClaims("/api/v1/courses/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      course_actual := &model.Course{}
      err := json.NewDecoder(w.Body).Decode(course_actual)
      g.Assert(err).Equal(nil)

      course_expected, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(course_actual.ID).Equal(course_expected.ID)
      g.Assert(course_actual.Name).Equal(course_expected.Name)
      g.Assert(course_actual.Description).Equal(course_expected.Description)
      g.Assert(course_actual.BeginsAt.Equal(course_expected.BeginsAt)).Equal(true)
      g.Assert(course_actual.EndsAt.Equal(course_expected.EndsAt)).Equal(true)
      g.Assert(course_actual.RequiredPercentage).Equal(course_expected.RequiredPercentage)
    })

    g.It("Should be able to filter enrollments (all)", func() {
      course_active, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      number_enrollments_expected, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1",
        course_active.ID,
      )
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/enrollments", 1, true)
      enrollments_actual := []model.UserCourse{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(number_enrollments_expected)
    })

    g.It("Should be able to filter enrollments (students only)", func() {
      course_active, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      number_enrollments_expected, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
        course_active.ID,
      )
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/enrollments?roles=0", 1, true)
      enrollments_actual := []model.UserCourse{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(number_enrollments_expected)
    })

    g.It("Should be able to filter enrollments (tutors only)", func() {
      course_active, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      number_enrollments_expected, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1 and role = 1",
        course_active.ID,
      )
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/enrollments?roles=1", 1, true)
      enrollments_actual := []model.UserCourse{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(number_enrollments_expected)
    })

    g.It("Should be able to filter enrollments (students+tutors only)", func() {
      course_active, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      number_enrollments_expected, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1 and role IN (0,1)",
        course_active.ID,
      )
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/enrollments?roles=0,1", 1, true)
      enrollments_actual := []model.UserCourse{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(number_enrollments_expected)
    })

    g.It("Should be able to filter enrollments (but receive only tutors + admins), when role=student", func() {
      course_active, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      number_enrollments_expected, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1 and role IN (1, 2)",
        course_active.ID,
      )
      g.Assert(err).Equal(nil)

      // 112 is a student
      w := tape.GetWithClaims("/api/v1/courses/1/enrollments?roles=0", 112, true)
      enrollments_actual := []model.UserCourse{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(number_enrollments_expected)
    })

    g.It("Creating course should require claims", func() {
      w := tape.Post("/api/v1/courses", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Creating course should require body", func() {
      w := tape.PlayWithClaims("POST", "/api/v1/courses", 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should create valid course", func() {

      courses_before, err := stores.Course.GetAll()
      g.Assert(err).Equal(nil)

      entry_sent := model.Course{
        Name:               "Info2_new",
        Description:        "Lorem Ipsum_new",
        BeginsAt:           helper.Time(time.Now()),
        EndsAt:             helper.Time(time.Now().Add(time.Hour * 1)),
        RequiredPercentage: 43,
      }

      // students
      w := tape.PlayDataWithClaims("POST", "/api/v1/courses", tape.ToH(entry_sent), 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PlayDataWithClaims("POST", "/api/v1/courses", tape.ToH(entry_sent), 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin in course (cannot be admin, course does not exists yet)
      w = tape.PlayDataWithClaims("POST", "/api/v1/courses", tape.ToH(entry_sent), 1, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PlayDataWithClaims("POST", "/api/v1/courses", tape.ToH(entry_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      // verify body
      course_return := &model.Course{}
      err = json.NewDecoder(w.Body).Decode(&course_return)
      g.Assert(course_return.Name).Equal(entry_sent.Name)
      g.Assert(course_return.Description).Equal(entry_sent.Description)
      g.Assert(course_return.BeginsAt.Equal(entry_sent.BeginsAt)).Equal(true)
      g.Assert(course_return.EndsAt.Equal(entry_sent.EndsAt)).Equal(true)
      g.Assert(course_return.RequiredPercentage).Equal(entry_sent.RequiredPercentage)

      // verify database
      course_new, err := stores.Course.Get(course_return.ID)
      g.Assert(err).Equal(nil)
      g.Assert(course_return.Name).Equal(course_new.Name)
      g.Assert(course_return.Description).Equal(course_new.Description)
      g.Assert(course_return.BeginsAt.Equal(course_new.BeginsAt)).Equal(true)
      g.Assert(course_return.EndsAt.Equal(course_new.EndsAt)).Equal(true)
      g.Assert(course_return.RequiredPercentage).Equal(course_new.RequiredPercentage)

      courses_after, err := stores.Course.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(courses_after)).Equal(len(courses_before) + 1)
    })

    g.Xit("Should send email to all enrolled users", func() {})

    g.It("Changes should require access claims", func() {
      w := tape.Put("/api/v1/courses/1", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should perform updates", func() {

      entry_sent := model.Course{
        Name:               "Info2_update",
        Description:        "Lorem Ipsum_update",
        BeginsAt:           helper.Time(time.Now()),
        EndsAt:             helper.Time(time.Now()),
        RequiredPercentage: 99,
      }

      // students
      w := tape.PlayDataWithClaims("PUT", "/api/v1/courses/1", tape.ToH(entry_sent), 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PlayDataWithClaims("PUT", "/api/v1/courses/1", tape.ToH(entry_sent), 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PlayDataWithClaims("PUT", "/api/v1/courses/1", tape.ToH(entry_sent), 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.Name).Equal(entry_sent.Name)
      g.Assert(entry_after.Description).Equal(entry_sent.Description)
      g.Assert(entry_after.BeginsAt.Equal(entry_sent.BeginsAt)).Equal(true)
      g.Assert(entry_after.EndsAt.Equal(entry_sent.EndsAt)).Equal(true)
      g.Assert(entry_after.RequiredPercentage).Equal(entry_sent.RequiredPercentage)
    })

    g.It("Should delete when valid access claims", func() {
      entries_before, err := stores.Course.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/courses/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // verify nothing has changes
      entries_after, err := stores.Course.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      // students
      w = tape.PlayWithClaims("DELETE", "/api/v1/courses/1", 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PlayWithClaims("DELETE", "/api/v1/courses/1", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PlayWithClaims("DELETE", "/api/v1/courses/1", 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a course less exists
      entries_after, err = stores.Course.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)
    })

    g.It("Can disenroll from course", func() {

      courseID := int64(1)

      number_enrollments_before, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
        courseID,
      )
      g.Assert(err).Equal(nil)

      w := tape.DeleteWithClaims("/api/v1/courses/1/enrollments", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      number_enrollments_after, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
        courseID,
      )
      g.Assert(err).Equal(nil)
      g.Assert(number_enrollments_after).Equal(number_enrollments_before - 1)

    })
    g.Xit("Cannot disenroll as a tutor from course", func() {})

    g.It("Permission test", func() {
      url := "/api/v1/courses/1"

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
