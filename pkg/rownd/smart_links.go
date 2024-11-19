package rownd

type SmartLinkOptions struct {
    Email       string                 `json:"email,omitempty"`
    Phone       string                 `json:"phone,omitempty"`
    RedirectURL string                 `json:"redirect_url"`
    Data        map[string]interface{} `json:"data,omitempty"`
}

type SmartLink struct {
    Link      string `json:"link"`
    AppUserID string `json:"app_user_id"`
}

func (c *Client) CreateSmartLink(opts *SmartLinkOptions) (*SmartLink, error) {
    resp, err := c.httpClient.DoRequest(
        context.Background(),
        "POST",
        fmt.Sprintf("%s/hub/auth/magic", c.baseURL),
        opts,
        &RequestOptions{
            Headers: map[string]string{
                "x-rownd-app-key":    c.appKey,
                "x-rownd-app-secret": c.appSecret,
            },
        },
    )
    if err != nil {
        return nil, err
    }

    var link SmartLink
    if err := DecodeResponse(resp, &link); err != nil {
        return nil, err
    }

    return &link, nil
}