package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
)

func main() {
    client, err := rownd.NewClient(&rownd.ClientConfig{
        AppKey:    "your-app-key",
        AppSecret: "your-app-secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    appID := "your-app-id"

    // Create a new group
    group, err := client.CreateGroup(ctx, appID, &rownd.CreateGroupRequest{
        Name:            "Engineering Team",
        AdmissionPolicy: "invite_only",
        Meta: map[string]interface{}{
            "department": "Engineering",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created group: %+v\n", group)

    // Create a group invite
    invite, err := client.CreateGroupInvite(ctx, appID, group.ID, &rownd.CreateGroupInviteRequest{
        Email:       "engineer@company.com",
        Roles:       []string{"member"},
        RedirectURL: "/welcome",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created invite: %+v\n", invite)

    // Add a member directly
    member, err := client.CreateGroupMember(ctx, appID, group.ID, &rownd.CreateGroupMemberRequest{
        UserID: "user_123",
        Roles:  []string{"admin", "member"},
        State:  "active",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Added member: %+v\n", member)

    // List all members
    members, err := client.ListGroupMembers(ctx, appID, group.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Group members: %+v\n", members)

    // Update member roles
    updatedMember, err := client.UpdateGroupMember(ctx, appID, group.ID, member.ID, &rownd.CreateGroupMemberRequest{
        UserID: member.UserID,
        Roles:  []string{"member"},
        State:  "active",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Updated member: %+v\n", updatedMember)

    // List all invites
    invites, err := client.ListGroupInvites(ctx, appID, group.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Group invites: %+v\n", invites)
}