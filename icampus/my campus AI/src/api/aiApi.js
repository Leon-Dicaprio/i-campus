// src/api/aiApi.js
export async function sendQuestion(question) {
  try {
    // 模拟网络延迟
    await new Promise(r => setTimeout(r, 600))

    const answers = [
      `关于"${question}"，建议您前往教务处官网查询相关规定。`,
      `您好！${question}？请使用校园APP中的“办事大厅”提交申请。`,
      `根据《学生手册》，${question}需提前3个工作日预约办理。`,
      `AI提示：${question}相关服务已上线“一站式服务中心”线上平台。`
    ]

    return { answer: answers[Math.floor(Math.random() * answers.length)] }
  } catch (error) {
    console.error('AI 请求失败:', error)
    return { answer: 'AI 暂时无法回答，请稍后再试。' }
  }
}