package main

import (
	"bufio"
	"fmt"
	"os"
	"quickflow/cmd/tech_ui/client"
	"quickflow/cmd/tech_ui/models"
	"quickflow/cmd/tech_ui/utils"
	"strings"
	"time"
)

func main() {
	// Choose server
	fmt.Println("Using local server (http://localhost:8080)")

	var baseURL string
	baseURL = "http://localhost:8080"

	apiClient := client.NewAPIClient(baseURL)

	for {
		fmt.Println("\n=== Social Network Client ===")
		fmt.Println("1. Login")
		fmt.Println("2. Sign Up")
		fmt.Println("3. Exit")

		choice := utils.ReadInt("Enter choice (1-3): ")

		switch choice {
		case 1:
			username := utils.ReadString("Username: ")
			password := utils.ReadString("Password: ")

			session, err := apiClient.Login(username, password)
			if err != nil {
				fmt.Printf("Login failed: %v\n", err)
				continue
			}

			fmt.Printf("Login successful! Session ID: %s\n", session.ID)
			mainMenu(apiClient)

		case 2:
			signUpForm := models.SignUpForm{
				Username:  utils.ReadString("Username: "),
				Password:  utils.ReadString("Password: "),
				Firstname: utils.ReadString("First name: "),
				Lastname:  utils.ReadString("Last name: "),
				BirthDate: utils.ReadDate("Birth date").Format(time.DateOnly),
				Sex:       utils.ReadInt("Sex (1-male, 2-female): "),
			}

			resp, err := apiClient.SignUp(signUpForm)
			if err != nil {
				fmt.Printf("Sign up failed: %v\n", err)
				continue
			}

			fmt.Printf("Sign up successful! User ID: %s\n", resp.ID)
			mainMenu(apiClient)

		case 3:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func mainMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Main Menu ===")
		fmt.Println("1. Users")
		fmt.Println("2. Posts")
		fmt.Println("3. Comments")
		fmt.Println("4. Likes")
		fmt.Println("5. Friends")
		fmt.Println("6. Communities")
		fmt.Println("7. Chats")
		//fmt.Println("8. Files")
		fmt.Println("8. WebSocket")
		fmt.Println("9. Logout")

		choice := utils.ReadInt("Enter choice (1-9): ")

		switch choice {
		case 1:
			usersMenu(apiClient)
		case 2:
			postsMenu(apiClient)
		case 3:
			commentsMenu(apiClient)
		case 4:
			likesMenu(apiClient)
		case 5:
			friendsMenu(apiClient)
		case 6:
			communitiesMenu(apiClient)
		case 7:
			chatsMenu(apiClient)
		//case 8:
		//	filesMenu(apiClient)
		case 8:
			websocketMenu(apiClient)
		case 9:
			apiClient.Logout()
			fmt.Println("Logged out")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func usersMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Users Menu ===")
		fmt.Println("1. Search Users")
		fmt.Println("2. Get Profile")
		fmt.Println("3. Update Profile")
		fmt.Println("4. Back")

		choice := utils.ReadInt("Enter choice (1-4): ")

		switch choice {
		case 1:
			query := utils.ReadString("Search query: ")
			count := utils.ReadInt("Count: ")

			users, err := apiClient.SearchUsers(query, count)
			if err != nil {
				fmt.Printf("Search failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d users:\n", len(users))
			for _, user := range users {
				fmt.Printf("- %s (%s %s) -- %s\n", user.Username, user.Firstname, user.Lastname, user.ID)
			}

		case 2:
			username := utils.ReadString("Username: ")

			profile, err := apiClient.GetProfile(username)
			if err != nil {
				fmt.Printf("Get profile failed: %v\n", err)
				continue
			}

			utils.PrintJSON(profile)

		case 3:
			firstname := utils.ReadString("First name: ")
			lastname := utils.ReadString("Last name: ")
			bio := utils.ReadString("Bio: ")
			birthDate := utils.ReadDate("Birth date").Format(time.DateOnly)
			sex := utils.ReadInt("Sex (1-male, 2-female): ")

			err := apiClient.UpdateProfile(firstname, lastname, bio, birthDate, sex)
			if err != nil {
				fmt.Printf("Update failed: %v\n", err)
				continue
			}

			fmt.Println("Profile updated successfully")

		case 4:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func postsMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Posts Menu ===")
		fmt.Println("1. Get User Posts")
		fmt.Println("2. Get Feed")
		fmt.Println("3. Get Post")
		fmt.Println("4. Create Post")
		fmt.Println("5. Update Post")
		fmt.Println("6. Delete Post")
		fmt.Println("7. Back")

		choice := utils.ReadInt("Enter choice (1-7): ")

		switch choice {
		case 1:
			username := utils.ReadString("Username: ")
			count := utils.ReadInt("Count: ")

			posts, err := apiClient.GetUserPosts(username, count, "")
			if err != nil {
				fmt.Printf("Get posts failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d posts:\n", len(posts))
			for _, post := range posts {
				fmt.Printf("- %s (post_id: %s): %s\n", post.Author.Username, post.ID, post.Text)
			}

		case 2:
			count := utils.ReadInt("Count: ")
			feedType := utils.ReadString("Type (feed/recommendations): ")

			posts, err := apiClient.GetFeed(count, "", feedType)
			if err != nil {
				fmt.Printf("Get feed failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d posts:\n", len(posts))
			for _, post := range posts {
				fmt.Printf("- %s (post_id: %s): %s\n", post.Author.Username, post.ID, post.Text)
			}

		case 3:
			postID := utils.ReadString("Post ID: ")

			post, err := apiClient.GetPost(postID)
			if err != nil {
				fmt.Printf("Get post failed: %v\n", err)
				continue
			}

			utils.PrintJSON(post)

		case 4:
			text := utils.ReadString("Text: ")

			form := models.CommentForm{
				Text: text,
			}

			post, err := apiClient.CreatePost(form)
			if err != nil {
				fmt.Printf("Create post failed: %v\n", err)
				continue
			}

			fmt.Printf("Post created: %s\n", post.ID)

		case 5:
			postID := utils.ReadString("Post ID: ")
			text := utils.ReadString("New text: ")

			form := models.UpdatePostForm{
				Text: text,
			}

			err := apiClient.UpdatePost(postID, form)
			if err != nil {
				fmt.Printf("Update post failed: %v\n", err)
				continue
			}

			fmt.Println("Post updated successfully")

		case 6:
			postID := utils.ReadString("Post ID: ")

			err := apiClient.DeletePost(postID)
			if err != nil {
				fmt.Printf("Delete post failed: %v\n", err)
				continue
			}

			fmt.Println("Post deleted successfully")

		case 7:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func websocketMenu(apiClient *client.APIClient) {
	fmt.Println("Connecting to WebSocket...")

	wsClient, err := apiClient.ConnectWebSocket()
	if err != nil {
		fmt.Printf("WebSocket connection failed: %v\n", err)
		return
	}
	defer wsClient.Close()

	fmt.Println("WebSocket connected successfully!")

	// Set up message handler
	wsClient.OnMessage(func(messageType string, payload []byte) {
		fmt.Printf("\n[WebSocket %s]: %s\n", messageType, string(payload))
	})

	// Interactive WebSocket menu
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n=== WebSocket Menu ===")
		fmt.Println("1. Send Message")
		fmt.Println("2. Mark Message Read")
		fmt.Println("3. Delete Message")
		fmt.Println("4. Delete Chat")
		fmt.Println("5. Back")

		fmt.Print("Enter choice (1-5): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			text := utils.ReadString("Message text: ")
			chatID := utils.ReadString("Chat ID (optional): ")
			receiverID := utils.ReadString("Receiver ID (optional): ")

			err := wsClient.SendTextMessage(text, chatID, receiverID, nil, nil, nil)
			if err != nil {
				fmt.Printf("Send message failed: %v\n", err)
			} else {
				fmt.Println("Message sent")
			}

		case "2":
			chatID := utils.ReadString("Chat ID: ")
			messageID := utils.ReadString("Message ID: ")

			err := wsClient.MarkMessageRead(chatID, messageID)
			if err != nil {
				fmt.Printf("Mark read failed: %v\n", err)
			} else {
				fmt.Println("Message marked as read")
			}

		case "3":
			messageID := utils.ReadString("Message ID: ")

			err := wsClient.DeleteMessage(messageID)
			if err != nil {
				fmt.Printf("Delete message failed: %v\n", err)
			} else {
				fmt.Println("Message deleted")
			}

		case "4":
			chatID := utils.ReadString("Chat ID: ")

			err := wsClient.DeleteChat(chatID)
			if err != nil {
				fmt.Printf("Delete chat failed: %v\n", err)
			} else {
				fmt.Println("Chat deleted")
			}

		case "5":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

// Добавьте эти функции после usersMenu в main.go

func commentsMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Comments Menu ===")
		fmt.Println("1. Get Post Comments")
		fmt.Println("2. Create Comment")
		fmt.Println("3. Update Comment")
		fmt.Println("4. Delete Comment")
		fmt.Println("5. Back")

		choice := utils.ReadInt("Enter choice (1-5): ")

		switch choice {
		case 1:
			postID := utils.ReadString("Post ID: ")
			count := utils.ReadInt("Count: ")

			comments, err := apiClient.GetPostComments(postID, count, "")
			if err != nil {
				fmt.Printf("Get comments failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d comments:\n", len(comments))
			for _, comment := range comments {
				fmt.Printf("- %s (comment_id: %s): %s\n", comment.Author.Username, comment.ID, comment.Text)
			}

		case 2:
			postID := utils.ReadString("Post ID: ")
			text := utils.ReadString("Comment text: ")

			form := models.CommentForm{
				Text: text,
			}

			comment, err := apiClient.CreateComment(postID, form)
			if err != nil {
				fmt.Printf("Create comment failed: %v\n", err)
				continue
			}

			fmt.Printf("Comment created: %s\n", comment.ID)

		case 3:
			postID := utils.ReadString("Post ID: ")
			commentID := utils.ReadString("Comment ID: ")
			text := utils.ReadString("New comment text: ")

			form := models.CommentForm{
				Text: text,
			}

			comment, err := apiClient.UpdateComment(postID, commentID, form)
			if err != nil {
				fmt.Printf("Update comment failed: %v\n", err)
				continue
			}

			fmt.Printf("Comment updated: %s\n", comment.ID)

		case 4:
			postID := utils.ReadString("Post ID: ")
			commentID := utils.ReadString("Comment ID: ")

			err := apiClient.DeleteComment(postID, commentID)
			if err != nil {
				fmt.Printf("Delete comment failed: %v\n", err)
				continue
			}

			fmt.Println("Comment deleted successfully")

		case 5:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func likesMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Likes Menu ===")
		fmt.Println("1. Like Post")
		fmt.Println("2. Like Comment")
		fmt.Println("3. Unlike Post")
		fmt.Println("4. Unlike Comment")
		fmt.Println("5. Back")

		choice := utils.ReadInt("Enter choice (1-5): ")

		switch choice {
		case 1:
			postID := utils.ReadString("Post ID: ")

			_, err := apiClient.LikePost(postID)
			if err != nil {
				fmt.Printf("Like post failed: %v\n", err)
				continue
			}

			fmt.Printf("Post liked\n")
			//Post ID: a000d8a1-b7a1-4d73-a482-b2ce942b1749
			//Comment ID: 866f5975-74b0-4f92-a86b-088661b1fea1
		case 2:
			postID := utils.ReadString("Post ID: ")
			commentID := utils.ReadString("Comment ID: ")

			_, err := apiClient.LikeComment(postID, commentID)
			if err != nil {
				fmt.Printf("Like comment failed: %v\n", err)
				continue
			}

			fmt.Printf("Comment liked\n")

		case 3:
			postID := utils.ReadString("Post ID: ")

			err := apiClient.UnlikePost(postID)
			if err != nil {
				fmt.Printf("Unlike post failed: %v\n", err)
				continue
			}

			fmt.Println("Post unliked successfully")

		case 4:
			postID := utils.ReadString("Post ID: ")
			commentID := utils.ReadString("Comment ID: ")

			err := apiClient.UnlikeComment(postID, commentID)
			if err != nil {
				fmt.Printf("Unlike comment failed: %v\n", err)
				continue
			}

			fmt.Println("Comment unliked successfully")

		case 5:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func friendsMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Friends Menu ===")
		fmt.Println("1. Get Friends")
		fmt.Println("2. Send Friend Request")
		fmt.Println("3. Respond to Friend Request")
		fmt.Println("4. Delete Friend")
		fmt.Println("5. Delete Friend Request")
		fmt.Println("6. Back")

		choice := utils.ReadInt("Enter choice (1-6): ")

		switch choice {
		case 1:
			userID := utils.ReadString("User ID: ")
			requestType := utils.ReadString("Request type (friend/outcoming/incoming): ")
			count := utils.ReadInt("Count: ")
			//offset := utils.ReadInt("Offset: ")

			friends, err := apiClient.GetFriends(userID, requestType, count, 0)
			if err != nil {
				fmt.Printf("Get friends failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d friends:\n", len(friends))
			for _, friend := range friends {
				status := "offline"
				if friend.IsOnline {
					status = "online"
				}
				fmt.Printf("- %s (%s %s: %s) [%s]\n", friend.Username, friend.Firstname, friend.Lastname, friend.ID, status)
			}

		case 2:
			receiverID := utils.ReadString("Receiver User ID: ")

			requestID, err := apiClient.SendFriendRequest(receiverID)
			if err != nil {
				fmt.Printf("Send friend request failed: %v\n", err)
				continue
			}

			fmt.Printf("Friend request sent: %s\n", requestID)

		case 3:
			requestID := utils.ReadString("Friend Request ID: ")
			status := utils.ReadString("Status (accepted/rejected): ")

			err := apiClient.RespondToFriendRequest(requestID, status)
			if err != nil {
				fmt.Printf("Respond to friend request failed: %v\n", err)
				continue
			}

			fmt.Println("Friend request responded successfully")

		case 4:
			friendID := utils.ReadString("Friend ID: ")

			err := apiClient.DeleteFriend(friendID)
			if err != nil {
				fmt.Printf("Delete friend failed: %v\n", err)
				continue
			}

			fmt.Println("Friend deleted successfully")

		case 5:
			requestID := utils.ReadString("Friend Request ID: ")

			err := apiClient.DeleteFriendRequest(requestID)
			if err != nil {
				fmt.Printf("Delete friend request failed: %v\n", err)
				continue
			}

			fmt.Println("Friend request deleted successfully")

		case 6:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func communitiesMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Communities Menu ===")
		fmt.Println("1. Search Communities")
		fmt.Println("2. Get User Communities")
		fmt.Println("3. Get Community by ID")
		fmt.Println("4. Get Community by Name")
		fmt.Println("5. Create Community")
		fmt.Println("6. Delete Community")
		fmt.Println("7. Join Community")
		fmt.Println("8. Leave Community")
		fmt.Println("9. Get Community Members")
		fmt.Println("10. Get Community Posts")
		fmt.Println("11. Create Community Post")
		fmt.Println("12. Back")

		choice := utils.ReadInt("Enter choice (1-12): ")

		switch choice {
		case 1:
			query := utils.ReadString("Search query: ")
			count := utils.ReadInt("Count: ")

			communities, err := apiClient.SearchCommunities(query, count)
			if err != nil {
				fmt.Printf("Search communities failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d communities:\n", len(communities))
			for _, community := range communities {
				fmt.Printf("- %s (%s): %s\n", community.Community.Name, community.Community.Nickname, community.Community.Description)
			}

		case 2:
			username := utils.ReadString("Username: ")
			count := utils.ReadInt("Count: ")
			role := utils.ReadString("Role filter (optional): ")

			communities, err := apiClient.GetUserCommunities(username, count, "", role)
			if err != nil {
				fmt.Printf("Get user communities failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d communities:\n", len(communities))
			for _, community := range communities {
				fmt.Printf("- %s (%s) [%s]\n", community.Community.Name, community.Community.Nickname, community.Role)
			}

		case 3:
			communityID := utils.ReadString("Community ID: ")

			community, err := apiClient.GetCommunityByID(communityID)
			if err != nil {
				fmt.Printf("Get community failed: %v\n", err)
				continue
			}

			utils.PrintJSON(community)

		case 4:
			name := utils.ReadString("Community name: ")

			community, err := apiClient.GetCommunityByName(name)
			if err != nil {
				fmt.Printf("Get community failed: %v\n", err)
				continue
			}

			utils.PrintJSON(community)

		case 5:
			nickname := utils.ReadString("Community nickname: ")
			name := utils.ReadString("Community name: ")
			description := utils.ReadString("Description: ")

			community, err := apiClient.CreateCommunity(nickname, name, description)
			if err != nil {
				fmt.Printf("Create community failed: %v\n", err)
				continue
			}

			fmt.Printf("Community created: %s\n", community.ID)

		case 6:
			communityID := utils.ReadString("Community ID: ")

			err := apiClient.DeleteCommunity(communityID)
			if err != nil {
				fmt.Printf("Delete community failed: %v\n", err)
				continue
			}

			fmt.Println("Community deleted successfully")

		case 7:
			communityID := utils.ReadString("Community ID: ")

			err := apiClient.JoinCommunity(communityID)
			if err != nil {
				fmt.Printf("Join community failed: %v\n", err)
				continue
			}

			fmt.Println("Joined community successfully")

		case 8:
			communityID := utils.ReadString("Community ID: ")
			userID := utils.ReadString("Your User ID: ")

			err := apiClient.LeaveCommunity(communityID, userID)
			if err != nil {
				fmt.Printf("Leave community failed: %v\n", err)
				continue
			}

			fmt.Println("Left community successfully")

		case 9:
			communityID := utils.ReadString("Community ID: ")
			count := utils.ReadInt("Count: ")

			members, err := apiClient.GetCommunityMembers(communityID, count, "")
			if err != nil {
				fmt.Printf("Get community members failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d members:\n", len(members))
			for _, member := range members {
				status := "offline"
				if member.Online {
					status = "online"
				}
				fmt.Printf("- %s (%s) [%s] - %s\n", member.Username, member.Role, status, member.JoinedAt.Format("2006-01-02"))
			}

		case 10:
			name := utils.ReadString("Community name: ")
			count := utils.ReadInt("Count: ")

			posts, err := apiClient.GetCommunityPosts(name, count, "")
			if err != nil {
				fmt.Printf("Get community posts failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d posts:\n", len(posts))
			for _, post := range posts {
				fmt.Printf("- %s (post_id: %s): %s\n", post.Author.Username, post.ID, post.Text)
			}

		case 11:
			communityName := utils.ReadString("Community name: ")
			text := utils.ReadString("Post text: ")

			form := models.CommentForm{
				Text: text,
			}

			post, err := apiClient.CreateCommunityPost(communityName, form)
			if err != nil {
				fmt.Printf("Create community post failed: %v\n", err)
				continue
			}

			fmt.Printf("Community post created: %s\n", post.ID)

		case 12:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func chatsMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Chats Menu ===")
		fmt.Println("1. Get Chats")
		fmt.Println("2. Get Chat Messages")
		fmt.Println("3. Back")

		choice := utils.ReadInt("Enter choice (1-3): ")

		switch choice {
		case 1:
			count := utils.ReadInt("Count: ")

			chats, err := apiClient.GetChats(count, "")
			if err != nil {
				fmt.Printf("Get chats failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d chats:\n", len(chats))
			for _, chat := range chats {
				status := "offline"
				if chat.Online {
					status = "online"
				}
				fmt.Printf("- %s (chat_id: %s) (%s) [%s] - %d unread\n", chat.Name, chat.ID, chat.Username, status, chat.UnreadMessages)
				if chat.LastMessage.Text != "" {
					fmt.Printf("  Last: %s\n", chat.LastMessage.Text)
				}
			}

		case 2:
			chatID := utils.ReadString("Chat ID: ")
			count := utils.ReadInt("Count: ")

			messages, err := apiClient.GetChatMessages(chatID, count, "")
			if err != nil {
				fmt.Printf("Get chat messages failed: %v\n", err)
				continue
			}

			fmt.Printf("Found %d messages:\n", len(messages))
			for _, message := range messages {
				fmt.Printf("- [%s] %s: %s\n", message.CreatedAt.Format("15:04"), message.Sender.Username, message.Text)
			}

		case 3:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func filesMenu(apiClient *client.APIClient) {
	for {
		fmt.Println("\n=== Files Menu ===")
		fmt.Println("1. Upload Files")
		fmt.Println("2. Back")

		choice := utils.ReadInt("Enter choice (1-2): ")

		switch choice {
		case 1:
			fmt.Println("Enter file paths (one per line, empty line to finish):")

			var mediaPaths, audioPaths, filePaths []string
			scanner := bufio.NewScanner(os.Stdin)

			fmt.Println("Media files:")
			for {
				fmt.Print("Path: ")
				scanner.Scan()
				path := strings.TrimSpace(scanner.Text())
				if path == "" {
					break
				}
				mediaPaths = append(mediaPaths, path)
			}

			fmt.Println("Audio files:")
			for {
				fmt.Print("Path: ")
				scanner.Scan()
				path := strings.TrimSpace(scanner.Text())
				if path == "" {
					break
				}
				audioPaths = append(audioPaths, path)
			}

			fmt.Println("Other files:")
			for {
				fmt.Print("Path: ")
				scanner.Scan()
				path := strings.TrimSpace(scanner.Text())
				if path == "" {
					break
				}
				filePaths = append(filePaths, path)
			}

			if len(mediaPaths) == 0 && len(audioPaths) == 0 && len(filePaths) == 0 {
				fmt.Println("No files to upload")
				continue
			}

			response, err := apiClient.UploadFiles(mediaPaths, audioPaths, filePaths)
			if err != nil {
				fmt.Printf("Upload failed: %v\n", err)
				continue
			}

			fmt.Println("Files uploaded successfully:")
			if len(response.Media) > 0 {
				fmt.Println("Media URLs:")
				for _, url := range response.Media {
					fmt.Printf("  - %s\n", url)
				}
			}
			if len(response.Files) > 0 {
				fmt.Println("File URLs:")
				for _, url := range response.Files {
					fmt.Printf("  - %s\n", url)
				}
			}
			if len(response.Audio) > 0 {
				fmt.Println("Audio URLs:")
				for _, url := range response.Audio {
					fmt.Printf("  - %s\n", url)
				}
			}

		case 2:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}
