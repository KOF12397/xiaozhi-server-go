package mattermost  
  
import (  
    "context"  
    "fmt"  
    "xiaozhi-server-go/src/core/utils"  
)  
  
// FunctionHandler Mattermost函数处理器  
type FunctionHandler struct {  
    client *Client  
    logger *utils.Logger  
}  
  
// NewFunctionHandler 创建函数处理器  
func NewFunctionHandler(client *Client, logger *utils.Logger) *FunctionHandler {  
    return &FunctionHandler{  
        client: client,  
        logger: logger,  
    }  
}  
  
// SendToMattermost 发送消息到Mattermost (LLM函数调用)  
func (fh *FunctionHandler) SendToMattermost(ctx context.Context, args map[string]interface{}) (interface{}, error) {  
    channel, ok := args["channel"].(string)  
    if !ok {  
        fh.logger.Error("Mattermost函数调用缺少channel参数: %+v", args)  
        return nil, fmt.Errorf("缺少channel参数")  
    }  
      
    message, ok := args["message"].(string)  
    if !ok {  
        fh.logger.Error("Mattermost函数调用缺少message参数: %+v", args)  
        return nil, fmt.Errorf("缺少message参数")  
    }  
  
    fh.logger.Info("LLM调用Mattermost函数 - 频道: %s, 消息长度: %d", channel, len(message))  
      
    err := fh.client.SendMessage(ctx, channel, message)  
    if err != nil {  
        fh.logger.Error("Mattermost函数调用发送失败: %v, 频道: %s", err, channel)  
        return nil, fmt.Errorf("发送Mattermost消息失败: %v", err)  
    }  
      
    return map[string]interface{}{  
        "status":  "success",  
        "message": "消息已发送到Mattermost",  
        "channel": channel,  
    }, nil  
}  
  
// GetFunctionDefinition 获取函数定义  
func (fh *FunctionHandler) GetFunctionDefinition() map[string]interface{} {  
    return map[string]interface{}{  
        "name":        "send_to_mattermost",  
        "description": "发送消息到Mattermost频道",  
        "parameters": map[string]interface{}{  
            "type": "object",  
            "properties": map[string]interface{}{  
                "channel": map[string]interface{}{  
                    "type":        "string",  
                    "description": "目标频道ID或名称",  
                },  
                "message": map[string]interface{}{  
                    "type":        "string",  
                    "description": "要发送的消息内容",  
                },  
            },  
            "required": []string{"channel", "message"},  
        },  
    }  
}
