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

package database

import (
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/jmoiron/sqlx"
)

type SubmissionStore struct {
  db *sqlx.DB
}

func NewSubmissionStore(db *sqlx.DB) *SubmissionStore {
  return &SubmissionStore{
    db: db,
  }
}

func (s *SubmissionStore) Get(submissionID int64) (*model.Submission, error) {
  p := model.Submission{}
  err := s.db.Get(&p, `SELECT * FROM tasks WHERE id = $1 LIMIT 1;`, p.ID)
  return &p, err
}

func (s *SubmissionStore) GetByUserAndTask(userID int64, taskID int64) (*model.Submission, error) {
  p := model.Submission{}
  err := s.db.Get(&p,
    `SELECT * FROM tasks WHERE user_id = $1 AND task_id = $2 LIMIT 1;`,
    userID, taskID)
  return &p, err
}

// func (s *SubmissionStore) GetAll() ([]model.Task, error) {
//   p := []model.Task{}
//   err := s.db.Select(&p, "SELECT * FROM tasks;")
//   return p, err
// }

// func (s *SubmissionStore) Create(p *model.Task, sheetID int64) (*model.Task, error) {
//   // create Task
//   newID, err := Insert(s.db, "tasks", p)
//   if err != nil {
//     return nil, err
//   }

//   // get maximum order
//   var maxOrder int
//   err = s.db.Get(&maxOrder, "SELECT max(ordering) FROM task_sheet WHERE sheet_id = $1", sheetID)
//   if err != nil {
//     return nil, err
//   }

//   // now associate sheet with course
//   _, err = s.db.Exec(`INSERT INTO task_sheet
//     (id,task_id,sheet_id,ordering)
//     VALUES (DEFAULT, $1, $2, $3);`, newID, sheetID, maxOrder+1)
//   if err != nil {
//     return nil, err
//   }

//   return s.Get(newID)
// }

// func (s *SubmissionStore) Update(p *model.Task) error {
//   return Update(s.db, "tasks", p.ID, p)
// }

// func (s *SubmissionStore) Delete(taskID int64) error {
//   return Delete(s.db, "tasks", taskID)
// }

// func (s *SubmissionStore) TasksOfSheet(sheetID int64, only_active bool) ([]model.Task, error) {
//   p := []model.Task{}

//   // t.public_test_path, t.private_test_path,
//   err := s.db.Select(&p, `
//     SELECT
//       t.id, t.created_at, t.updated_at, t.max_points,
//       t.public_docker_image, t.private_docker_image
//     FROM task_sheet ts
//     INNER JOIN
//       tasks t ON ts.task_id = t.id
//     INNER JOIN
//       sheets s ON ts.sheet_id = s.id
//     WHERE
//       s.id = $1
//     ORDER BY
//       ts.ordering ASC;`, sheetID)
//   return p, err
// }