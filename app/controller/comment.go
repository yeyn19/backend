package controller

import (
	"encoding/json"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

// CreateComment creates a comment by father_id(0 for root node), target_id,target_type
func CreateComment(c echo.Context) error {
	req := request.CreateCommentRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	user := c.Get("user").(models.User)

	newReaction := models.Reaction{
		TargetType: "comment",
	}
	utils.PanicIfDBError(base.DB.Save(&newReaction), "could not save reaction")
	var newComment models.Comment

	if req.FatherID != 0 {
		//father is comment
		var father models.Comment
		base.DB.Model(&models.Comment{}).First(&father, uint(req.FatherID))
		newComment := models.Comment{
			User:       user,
			Reaction:   newReaction,
			FatherID:   uint(req.FatherID),
			TargetType: father.TargetType,
			TargetID:   father.TargetID,
			Content:    req.Content,
		}
		if father.FatherID != 0 {
			// father is comment node
			newComment.RootCommentID = father.RootCommentID
		} else {
			// father is root node
			newComment.RootCommentID = father.ID
		}
		utils.PanicIfDBError(base.DB.Save(&newComment), "could not save comment")

	} else {
		//root node
		newComment := models.Comment{
			User:       user,
			Reaction:   newReaction,
			TargetType: req.TargetType,
			TargetID:   uint(req.TargetID),
			Content:    req.Content,
		}
		utils.PanicIfDBError(base.DB.Save(&newComment), "could not save comment")
	}

	return c.JSON(http.StatusCreated, response.CreateCommentResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Comment models.Comment
		}{
			newComment,
		},
	})

}

// GetComment query a comment for a problem or a homework by id(uint) and type(string), returns rootNodes and noneRootNodes
func GetComment(c echo.Context) error {
	req := request.GetCommentRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	var NotRootComments []models.Comment
	var RootComments []models.Comment
	query := base.DB.Model(&models.Comment{}).
		Preload("User").
		Preload("Reaction").
		Order("ID").
		Where(" target_type = (?) AND target_id = (?) AND father_id = (?)", req.TargetType, uint(req.TargetID), 0)
	//paginator
	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &RootComments)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}

	//query roots' children, already been paginated
	var RecursiveFatherIds []uint
	for _, v := range RootComments {
		RecursiveFatherIds = append(RecursiveFatherIds, v.ID)
	}

	base.DB.Model(&models.Comment{}).
		Preload("User").
		Preload("Reaction").
		Order("updated_at desc").
		Find(&NotRootComments, "root_comment_id in ?", RecursiveFatherIds)

	return c.JSON(http.StatusCreated, response.GetCommentResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			RootComments    []models.Comment
			NotRootComments []models.Comment
			Total           int     `json:"total"`
			Offset          int     `json:"offset"`
			Prev            *string `json:"prev"`
			Next            *string `json:"next"`
		}{
			RootComments,
			NotRootComments,
			total,
			req.Offset,
			prevUrl,
			nextUrl,
		},
	})
}

// AddReaction makes a reaction, assert frontend have checked if the operation is illegaled
func AddReaction(c echo.Context) error {
	req := request.AddReactionRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	user := c.Get("user").(models.User)

	if req.TargetType == "comment" {
		var comment models.Comment
		base.DB.Model(&models.Comment{}).
			Preload("Reaction").
			First(&comment, uint(req.TargetID))
		var targetReaction models.Reaction
		if comment.ReactionID == 0 {
			// don't have reaction
			targetReaction = models.Reaction{
				TargetType: "comment",
			}
			comment.Reaction = targetReaction
		} else {
			targetReaction = comment.Reaction
		}

		maps := make(map[string][]uint)
		if targetReaction.Details == "" {
			maps[req.EmojiType] = make([]uint, 2)
			maps[req.EmojiType][0] = 1
			maps[req.EmojiType][1] = uint(user.ID)
		} else {
			err := json.Unmarshal([]byte(targetReaction.Details), &maps)
			if err != nil {
				panic(errors.Wrap(err, "could not marshal reaction map"))
			}
			_, key := maps[req.EmojiType]
			if key {
				//operator is not zero
				for _, v := range maps[req.EmojiType] {
					if uint(v) == uint(user.ID) {
						return c.JSON(http.StatusBadRequest, response.ErrorResp("can't action twice!", nil))
					}
				}
				maps[req.EmojiType] = append(maps[req.EmojiType], uint(user.ID))
				maps[req.EmojiType][0] += 1
			} else {
				maps[req.EmojiType] = make([]uint, 2)
				maps[req.EmojiType][0] = 1
				maps[req.EmojiType][1] = uint(user.ID)
			}
		}
		jsonStr, err := json.Marshal(maps)
		if err != nil {
			panic(errors.Wrap(err, "could not marshal reaction map"))
		}
		targetReaction.Details = string(jsonStr)
		utils.PanicIfDBError(base.DB.Save(&(targetReaction)), "save reaction error")

		if comment.ReactionID == 0 {
			comment.Reaction = targetReaction
			utils.PanicIfDBError(base.DB.Save(&comment), "save reaction error")
		}

	} else {
		// todo: implement this
	}

	return c.JSON(http.StatusCreated, response.AddReactionResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Content string
		}{
			"you have successfully " + req.EmojiType + "ed the comment",
		},
	})

}

// DeleteReaction deletes a reaction, assert frontend have checked if the operation is illegaled
func DeleteReaction(c echo.Context) error {
	req := request.DeleteReactionRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	user := c.Get("user").(models.User)

	if req.TargetType == "comment" {
		var comment models.Comment
		base.DB.Model(&models.Comment{}).
			Preload("Reaction").
			First(&comment, uint(req.TargetID))
		var targetReaction models.Reaction
		if comment.ReactionID == 0 {
			// don't have reaction
			targetReaction = models.Reaction{
				TargetType: "comment",
			}
			comment.Reaction = targetReaction
		} else {
			targetReaction = comment.Reaction
		}

		maps := make(map[string][]uint)
		//delete action
		if targetReaction.Details == "" {
			panic("this should not happen, logic in error!")
		} else {
			json.Unmarshal([]byte(targetReaction.Details), &maps)
			_, key := maps[req.EmojiType]
			if key {
				pos := -1
				for k, v := range maps[req.EmojiType] {
					//skip the first, standing for counts
					if k != 0 && v == uint(user.ID) {
						pos = k
						break
					}
				}
				if pos == -1 {
					panic("should not happen, you have not actived")
				} else {
					maps[req.EmojiType] = append(maps[req.EmojiType][:pos], maps[req.EmojiType][pos+1:]...)
					maps[req.EmojiType][0] -= 1
				}
				jsonStr, err := json.Marshal(maps)
				if err != nil {
					panic(errors.Wrap(err, "could not marshal reaction map"))
				}
				targetReaction.Details = string(jsonStr)
				utils.PanicIfDBError(base.DB.Save(&targetReaction), "save reaction error")
			} else {
				return c.JSON(http.StatusBadRequest, response.ErrorResp("you have not actived!", nil))
			}
		}

		if comment.ReactionID == 0 {
			comment.Reaction = targetReaction
			utils.PanicIfDBError(base.DB.Save(&comment), "save reaction error")
		}

	} else {
		// todo: implement this
	}

	return c.JSON(http.StatusCreated, response.AddReactionResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Content string
		}{
			"you have successfully " + req.EmojiType + "ed the comment",
		},
	})

}

// DeleteComment deletes a comment with id, and we have hook in database/models/comment.go to recursive delete it's children
func DeleteComment(c echo.Context) error {
	user, ok := c.Get("user").(models.User)
	if !ok {
		panic("could not convert my user into type models.User")
	}

	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(errors.Wrap(err, "could find target comment id"))
	}

	var comment models.Comment
	base.DB.Model(&models.Comment{}).
		Preload("Reaction").
		First(&comment, uint(commentID))

	//check whether has role to delete comemnt
	canDeleteComment := user.HasRole("admin") || (comment.UserID == user.ID)
	if canDeleteComment == false {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("you don't have permission to delete this comment", nil))
	}

	utils.PanicIfDBError(base.DB.Delete(&comment), "could not delete target comment")

	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})

}
