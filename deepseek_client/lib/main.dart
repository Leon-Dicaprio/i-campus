import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

void main() {
  runApp(const DeepSeekApp());
}

class DeepSeekApp extends StatelessWidget {
  const DeepSeekApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'DeepSeek MVP',
      theme: ThemeData(
        // DeepSeek 标志性的深蓝色调
        primaryColor: const Color(0xFF1B2432),
        scaffoldBackgroundColor: const Color(0xFFF7F9FB),
        useMaterial3: true,
      ),
      home: const ChatScreen(),
    );
  }
}

// 消息模型
class ChatMessage {
  final String text;
  final bool isUser;
  ChatMessage({required this.text, required this.isUser});
}

class ChatScreen extends StatefulWidget {
  const ChatScreen({super.key});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final TextEditingController _controller = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [];
  bool _isLoading = false;

  // 核心：发送消息并处理流式响应
  Future<void> _sendMessage() async {
    final text = _controller.text.trim();
    if (text.isEmpty) return;

    setState(() {
      _messages.add(ChatMessage(text: text, isUser: true));
      _messages.add(ChatMessage(text: "", isUser: false)); // 占位符，用于流式填充
      _isLoading = true;
      _controller.clear();
    });
    _scrollToBottom();

    try {
      // 创建请求
      final request = http.Request('POST', Uri.parse('http://10.0.2.2:8080/api/chat'));
      request.headers['Content-Type'] = 'application/json';
      request.body = jsonEncode({"message": text});

      // 发送请求并获取流 (StreamedResponse)
      final response = await request.send();

      // 监听数据流
      response.stream
          .transform(utf8.decoder) // 解码二进制流
          .transform(const LineSplitter()) // 按行分割
          .listen((line) {
            if (line.startsWith("data: ")) {
              // 提取 SSE 数据内容
              final content = line.substring(6); 
              setState(() {
                // 将新字符追加到最后一条消息（即 AI 的回复）
                final lastMsg = _messages.last;
                _messages[_messages.length - 1] = ChatMessage(
                  text: lastMsg.text + content, 
                  isUser: false
                );
              });
              _scrollToBottom();
            }
          }, onDone: () {
            setState(() => _isLoading = false);
          }, onError: (e) {
            setState(() {
               _messages.add(ChatMessage(text: "Error: $e", isUser: false));
               _isLoading = false;
            });
          });
          
    } catch (e) {
      setState(() => _isLoading = false);
      // 简单错误处理
    }
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text("DeepSeek", style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
        backgroundColor: const Color(0xFF4D6BFE), // DeepSeek 品牌蓝
        elevation: 0,
      ),
      body: Column(
        children: [
          // 消息列表区
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                final msg = _messages[index];
                return Align(
                  alignment: msg.isUser ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: const EdgeInsets.symmetric(vertical: 8),
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: msg.isUser ? const Color(0xFF4D6BFE) : Colors.white,
                      borderRadius: BorderRadius.circular(12),
                      boxShadow: [
                        if (!msg.isUser)
                          BoxShadow(color: Colors.grey.withOpacity(0.1), blurRadius: 4, spreadRadius: 1)
                      ]
                    ),
                    child: Text(
                      msg.text,
                      style: TextStyle(
                        color: msg.isUser ? Colors.white : Colors.black87,
                        fontSize: 16,
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          // 底部输入区
          Container(
            padding: const EdgeInsets.all(16),
            decoration: const BoxDecoration(
              color: Colors.white,
              border: Border(top: BorderSide(color: Color(0xFFEEEEEE))),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _controller,
                    decoration: InputDecoration(
                      hintText: "给 DeepSeek 发送消息",
                      filled: true,
                      fillColor: const Color(0xFFF2F4F7),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(24),
                        borderSide: BorderSide.none,
                      ),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                    ),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                const SizedBox(width: 8),
                IconButton(
                  onPressed: _isLoading ? null : _sendMessage,
                  icon: Icon(
                    Icons.arrow_upward, 
                    color: _isLoading ? Colors.grey : const Color(0xFF4D6BFE)
                  ),
                  style: IconButton.styleFrom(
                    backgroundColor: const Color(0xFFE8EBFF),
                    shape: const CircleBorder(),
                  ),
                )
              ],
            ),
          )
        ],
      ),
    );
  }
}