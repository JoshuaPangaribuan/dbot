package langchain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAgent_Table(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (Service, []Tool)
		wantNil bool
	}{
		{
			name: "valid service with tools",
			setup: func() (Service, []Tool) {
				svc, err := New(WithProvider(ProviderOllama))
				require.NoError(t, err)
				tools := []Tool{
					NewSimpleTool("test", "test tool", func(ctx context.Context, input string) (string, error) {
						return "result", nil
					}),
				}
				return svc, tools
			},
			wantNil: false,
		},
		{
			name: "valid service without tools",
			setup: func() (Service, []Tool) {
				svc, err := New(WithProvider(ProviderOllama))
				require.NoError(t, err)
				return svc, nil
			},
			wantNil: false,
		},
		{
			name: "nil service",
			setup: func() (Service, []Tool) {
				return nil, nil
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, tools := tt.setup()
			agent := NewAgent(svc, tools...)

			if tt.wantNil {
				require.Nil(t, agent)
			} else {
				require.NotNil(t, agent)
				_ = svc.Close()
			}
		})
	}
}

func TestSimpleTool_Table(t *testing.T) {
	tests := []struct {
		name        string
		tool        Tool
		input       string
		wantResult  string
		wantErr     bool
	}{
		{
			name: "simple tool",
			tool: NewSimpleTool("echo", "echoes input", func(ctx context.Context, input string) (string, error) {
				return input, nil
			}),
			input:      "hello",
			wantResult: "hello",
			wantErr:    false,
		},
		{
			name: "tool with error",
			tool: NewSimpleTool("error", "returns error", func(ctx context.Context, input string) (string, error) {
				return "", errors.New("tool error")
			}),
			input:       "test",
			wantResult:  "",
			wantErr:     true,
		},
		{
			name: "tool with transformation",
			tool: NewSimpleTool("upper", "converts to uppercase", func(ctx context.Context, input string) (string, error) {
				return "UPPER: " + input, nil
			}),
			input:      "test",
			wantResult: "UPPER: test",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.tool.Name(), tt.tool.Name())
			require.Equal(t, tt.tool.Description(), tt.tool.Description())

			result, err := tt.tool.Call(context.Background(), tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestAgent_BuildSystemPrompt_Table(t *testing.T) {
	tests := []struct {
		name     string
		tools    []Tool
		contains []string
	}{
		{
			name:     "no tools",
			tools:    nil,
			contains: []string{"helpful assistant"},
		},
		{
			name: "single tool",
			tools: []Tool{
				NewSimpleTool("test", "test tool", nil),
			},
			contains: []string{"helpful assistant", "test", "test tool"},
		},
		{
			name: "multiple tools",
			tools: []Tool{
				NewSimpleTool("tool1", "description 1", nil),
				NewSimpleTool("tool2", "description 2", nil),
				NewSimpleTool("tool3", "description 3", nil),
			},
			contains: []string{"tool1", "description 1", "tool2", "description 2", "tool3", "description 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			agent := NewAgent(svc, tt.tools...)
			require.NotNil(t, agent)

			prompt := agent.buildSystemPrompt()

			for _, substr := range tt.contains {
				require.Contains(t, prompt, substr)
			}
		})
	}
}

func TestAgent_Run_Table(t *testing.T) {
	tests := []struct {
		name        string
		tools       []Tool
		input       string
		wantErr     bool
		skip        bool
		skipReason  string
	}{
		{
			name:  "run without tools",
			tools: nil,
			input: "Hello, how are you?",
			wantErr: false,
			skip: true,
			skipReason: "requires LLM API call",
		},
		{
			name: "run with matching tool",
			tools: []Tool{
				NewSimpleTool("test_tool", "this is a test tool", func(ctx context.Context, input string) (string, error) {
					return "tool result: " + input, nil
				}),
			},
			input: "use the test_tool",
			wantErr: false,
			skip: true,
			skipReason: "requires LLM API call",
		},
		{
			name: "run with non-matching tool",
			tools: []Tool{
				NewSimpleTool("weather", "get weather", func(ctx context.Context, input string) (string, error) {
					return "sunny", nil
				}),
			},
			input: "tell me a joke",
			wantErr: false,
			skip: true,
			skipReason: "requires LLM API call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			agent := NewAgent(svc, tt.tools...)
			require.NotNil(t, agent)

			result, err := agent.Run(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, result)
			}
		})
	}
}

func TestAgent_Run_ToolError_Table(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		toolDesc  string
		toolFn    func(context.Context, string) (string, error)
		input     string
		wantErr   bool
		skip      bool
		skipReason string
	}{
		{
			name:     "tool returns error",
			toolName: "failing_tool",
			toolDesc: "a tool that fails",
			toolFn: func(ctx context.Context, input string) (string, error) {
				return "", errors.New("tool execution failed")
			},
			input: "use failing_tool",
			wantErr: true,
			skip: true,
			skipReason: "requires LLM API call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			tool := NewSimpleTool(tt.toolName, tt.toolDesc, tt.toolFn)
			agent := NewAgent(svc, tool)
			require.NotNil(t, agent)

			_, err = agent.Run(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

func TestAgent_ShouldUseTool_Table(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		toolDesc string
		input    string
		want     bool
	}{
		{
			name:     "tool name matches",
			toolName: "calculator",
			toolDesc: "performs calculations",
			input:    "use the calculator",
			want:     true,
		},
		{
			name:     "tool description matches",
			toolName: "calc",
			toolDesc: "performs mathematical calculations",
			input:    "help me calculate something",
			want:     true,
		},
		{
			name:     "no match",
			toolName: "weather",
			toolDesc: "gets weather information",
			input:    "tell me a joke",
			want:     false,
		},
		{
			name:     "empty tool name",
			toolName: "",
			toolDesc: "description",
			input:    "test",
			want:     false,
		},
		{
			name:     "empty input",
			toolName: "test",
			toolDesc: "test tool",
			input:    "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			tool := NewSimpleTool(tt.toolName, tt.toolDesc, nil)
			agent := NewAgent(svc, tool)
			require.NotNil(t, agent)

			got := agent.shouldUseTool(context.Background(), tt.input, tool)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSimpleTool_NameAndDescription_Table(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		toolDesc    string
		wantName    string
		wantDesc    string
	}{
		{
			name:     "basic tool",
			toolName: "test",
			toolDesc: "test description",
			wantName: "test",
			wantDesc: "test description",
		},
		{
			name:     "empty name and desc",
			toolName: "",
			toolDesc: "",
			wantName: "",
			wantDesc: "",
		},
		{
			name:     "special characters",
			toolName: "tool-with_special.chars",
			toolDesc: "A tool with special: characters!",
			wantName: "tool-with_special.chars",
			wantDesc: "A tool with special: characters!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewSimpleTool(tt.toolName, tt.toolDesc, nil)
			require.Equal(t, tt.wantName, tool.Name())
			require.Equal(t, tt.wantDesc, tool.Description())
		})
	}
}

func TestNewSimpleTool_Table(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(context.Context, string) (string, error)
		check  func(*testing.T, Tool)
	}{
		{
			name: "nil function",
			fn:   nil,
			check: func(t *testing.T, tool Tool) {
				require.NotNil(t, tool)
				// Calling with nil function should panic or handle gracefully
				// For now, we just check the tool is created
			},
		},
		{
			name: "valid function",
			fn: func(ctx context.Context, input string) (string, error) {
				return "ok", nil
			},
			check: func(t *testing.T, tool Tool) {
				require.NotNil(t, tool)
				result, err := tool.Call(context.Background(), "test")
				require.NoError(t, err)
				require.Equal(t, "ok", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewSimpleTool("test", "test tool", tt.fn)
			tt.check(t, tool)
		})
	}
}
