#!/bin/bash

# siti-cli Bash 补全脚本
# 自动发现和解析 commands 目录下的所有命令

_siti_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # 获取 siti 脚本所在目录
    local siti_script
    siti_script=$(which siti 2>/dev/null)
    if [ -z "$siti_script" ]; then
        return 0
    fi
    
    # 智能检测 commands 目录（兼容 Homebrew 和源码安装）
    local script_dir
    script_dir=$(dirname "$siti_script")
    local commands_dir
    
    # 检查是否通过 Homebrew 安装
    if [[ "$script_dir" == "/opt/homebrew/bin" ]] || [[ "$script_dir" == "/usr/local/bin" ]]; then
        # Homebrew 安装
        if [ -d "/opt/homebrew/share/siti-cli/commands" ]; then
            commands_dir="/opt/homebrew/share/siti-cli/commands"
        elif [ -d "/usr/local/share/siti-cli/commands" ]; then
            commands_dir="/usr/local/share/siti-cli/commands"
        else
            return 0
        fi
    else
        # 源码开发模式
        commands_dir="$(dirname "$script_dir")/src/commands"
    fi
    
    # 如果 commands 目录不存在，返回
    if [ ! -d "$commands_dir" ]; then
        return 0
    fi
    
    # 获取所有可用命令
    local commands=()
    for cmd_file in "$commands_dir"/*.sh; do
        if [ -f "$cmd_file" ]; then
            local cmd_name
            cmd_name=$(basename "$cmd_file" .sh)
            # 将下划线转换为连字符
            cmd_name="${cmd_name//_/-}"
            commands+=("$cmd_name")
        fi
    done
    
    # 全局选项
    local global_opts="--help --version -h -v"
    
    # 如果当前是第一个参数，显示所有命令和全局选项
    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands[*]} ${global_opts}" -- ${cur}) )
        return 0
    fi
    
    # 如果前一个参数是 --help 或 -h，不需要补全
    if [ "$prev" = "--help" ] || [ "$prev" = "-h" ]; then
        return 0
    fi
    
    # 获取当前命令（第一个参数）
    local current_cmd="${COMP_WORDS[1]}"
    
    # 将连字符转换回下划线
    current_cmd="${current_cmd//-/_}"
    local cmd_script="$commands_dir/${current_cmd}.sh"
    
    # 如果命令脚本不存在，返回
    if [ ! -f "$cmd_script" ]; then
        return 0
    fi
    
    # 从脚本中提取补全信息
    local completion_opts=""
    
    # 查找补全信息：以 "# 补全:" 开头的行
    while IFS= read -r line; do
        # 跳过注释符号和空格
        line=$(echo "$line" | sed 's/^# *//')
        
        # 提取选项和描述
        if [[ "$line" =~ ^([^:]+):(.+)$ ]]; then
            local option="${BASH_REMATCH[1]}"
            local desc="${BASH_REMATCH[2]}"
            # 清理选项名（去除前后空格）
            option=$(echo "$option" | xargs)
            completion_opts="$completion_opts $option"
        fi
    done < <(awk '/^# 补全:/,/^# [^ ]/ { if (/^# 补全:/) next; if (/^# [^ ]/ && !/^# 补全:/) exit; print }' "$cmd_script")
    
    # 如果找到了补全选项，使用它们
    if [ -n "$completion_opts" ]; then
        COMPREPLY=( $(compgen -W "${completion_opts}" -- ${cur}) )
    else
        # 如果没有找到补全信息，尝试从用法中提取
        local usage_opts=""
        while IFS= read -r line; do
            # 查找类似 "siti cmd option" 的模式
            if [[ "$line" =~ siti[[:space:]]+[^[:space:]]+[[:space:]]+([^[:space:]]+) ]]; then
                local option="${BASH_REMATCH[1]}"
                usage_opts="$usage_opts $option"
            fi
        done < <(awk '/^# 用法:/,/^# [^ ]/ { if (/^# 用法:/) next; if (/^# [^ ]/ && !/^# 用法:/) exit; print }' "$cmd_script")
        
        if [ -n "$usage_opts" ]; then
            COMPREPLY=( $(compgen -W "${usage_opts}" -- ${cur}) )
        fi
    fi
    
    return 0
}

# 注册补全函数
complete -F _siti_completion siti
