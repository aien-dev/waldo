// Copyright (c) 2026 OpenWALDO Project contributors
// Copyright (c) 2026 CtrlIQ, Inc.
// Copyright (c) 2026 Gregory M. Kurtzer
// SPDX-License-Identifier: Apache-2.0

package modelexport

import (
	"fmt"

	"github.com/openwaldo/waldo/internal/model"
)

const userAssistantJinja = `{%- set start = 0 %}
{%- if tools %}
{%- if messages and messages[0]['role'] == 'system' %}
{{- 'System: ' + messages[0]['content'] + '\n\nAvailable tools:\n' }}{{ tools | tojson }}
{%- set start = 1 %}
{%- else %}
{{- 'System: Available tools:\n' }}{{ tools | tojson }}
{%- endif %}
{%- endif %}
{%- for message in messages[start:] %}
{{- '\n\n' }}
{%- if message['role'] == 'assistant' %}
{{- 'Assistant:' + message['content'] }}{%- if message['tool_calls'] %}{{ message['tool_calls'] | tojson }}{%- endif %}
{%- elif message['role'] == 'tool' %}
{{- 'Tool: ' + message['content'] }}
{%- elif message['role'] == 'user' %}
{{- 'User: ' + message['content'] }}
{%- else %}
{{- 'System: ' + message['content'] }}
{%- endif %}
{%- endfor %}
{%- if add_generation_prompt %}{{ '\n\nAssistant:' }}{%- endif %}`

const userAssistantJinjaWithoutTools = `{%- for message in messages %}
{%- if not loop.first %}{{ '\n\n' }}{%- endif %}
{%- if message['role'] == 'assistant' %}
{{- 'Assistant:' + message['content'] }}
{%- elif message['role'] == 'user' %}
{{- 'User: ' + message['content'] }}
{%- else %}
{{- 'System: ' + message['content'] }}
{%- endif %}
{%- endfor %}
{%- if add_generation_prompt %}{{ '\n\nAssistant:' }}{%- endif %}`

const chatMLJinja = `{%- set start = 0 %}
{%- if tools %}
{{- '<|im_start|>system\n' }}
{%- if messages and messages[0]['role'] == 'system' %}
{{- messages[0]['content'] + '\n\n' }}
{%- set start = 1 %}
{%- endif %}
{{- 'Available tools:\n' }}{{ tools | tojson }}{{ '<|im_end|>\n' }}
{%- endif %}
{%- for message in messages[start:] %}
{{- '<|im_start|>' + message['role'] + '\n' + message['content'] }}
{%- if message['role'] == 'assistant' and message['tool_calls'] %}{{ message['tool_calls'] | tojson }}{%- endif %}
{{- '<|im_end|>\n' }}
{%- endfor %}
{%- if add_generation_prompt %}{{ '<|im_start|>assistant\n' }}{%- endif %}`

// chatMLJinjaWithoutTools is the plain ChatML layout in the exact text Hugging Face model repos
// ship for it (SmolLM2-Instruct and others without a default system message). Its newlines are
// real newline characters inside the Jinja string literals, as in those repos. It renders the same
// bytes as the previous whitespace-controlled spelling of this template; the text is fixed so that
// consumers that recognise the standard template by its text (for example the AIEN loader) accept
// the export.
const chatMLJinjaWithoutTools = "{% for message in messages %}{{'<|im_start|>' + message['role'] + '\n' + message['content'] + '<|im_end|>' + '\n'}}{% endfor %}{% if add_generation_prompt %}{{ '<|im_start|>assistant\n' }}{% endif %}"

func jinjaInteractionTemplate(interaction model.Interaction) (string, error) {
	if err := interaction.Validate(); err != nil {
		return "", err
	}
	if !interaction.Conversational() {
		return "", nil
	}
	switch interaction.Template {
	case model.InteractionUserAssistantV1:
		if !interaction.Tools {
			return userAssistantJinjaWithoutTools, nil
		}
		return userAssistantJinja, nil
	case model.InteractionChatMLV1:
		if !interaction.Tools {
			return chatMLJinjaWithoutTools, nil
		}
		return chatMLJinja, nil
	default:
		return "", fmt.Errorf("unsupported export interaction template %q", interaction.Template)
	}
}
