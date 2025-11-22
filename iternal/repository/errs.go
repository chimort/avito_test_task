package repository

import "errors"

var ErrTeamExists = errors.New("team already exists")
var ErrTeamNotFound = errors.New("team not found")
var ErrPRExists = errors.New("PR already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrPRNotFound = errors.New("PR not found")
var ErrPRMerged = errors.New("can not ressign on merged pr")
var ErrReviewerNotAssign = errors.New("no is not assigned")
var ErrNoCandidates = errors.New("no active replacement candidates")