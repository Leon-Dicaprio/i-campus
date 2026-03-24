<template>
  <div class="student-handbook">
    <div class="handbook-hero">
      <h2>东莞理工学院学生手册</h2>
      <p>内容总结与详细解读</p>
    </div>

    <div class="toc-card">
      <div class="toc-title">索引目录</div>
      <div class="toc-list">
        <button v-for="item in toc" :key="item.id" class="toc-item" @click="scrollTo(item.id)">
          {{ item.title }}
        </button>
      </div>
    </div>

    <div class="sections">
      <details class="section" open>
        <summary id="sec-1">一、学校核心理念与定位</summary>
        <div class="section-body">
          <h4>（一）核心精神理念</h4>
          <ul>
            <li><strong>校训</strong>：学而知不足</li>
            <li><strong>办学思想</strong>：知行合一，立德树人</li>
            <li><strong>莞工精神</strong>：爱国奋斗，追求卓越</li>
            <li><strong>战略理念</strong>：守正出新，登高致远</li>
            <li><strong>价值追求</strong>：以卓越的创新教育与实践造福社会</li>
          </ul>

          <h4>（二）人才培养相关</h4>
          <ul>
            <li><strong>培养目标</strong>：培养适应现代产业发展需求，勇于担当、善于学习、敢于超越的高素质应用型创新人才</li>
            <li>
              <strong>培养总标准</strong>：涵盖理想信念、社会责任感、职业道德、人文素养、沟通协作、终身学习、专业基础、实践能力、创新精神、国际化视野等九大维度
            </li>
          </ul>
        </div>
      </details>

      <details class="section">
        <summary id="sec-2">二、五大核心制度板块</summary>
        <div class="section-body">
          <details class="sub-section" open>
            <summary>（一）国家规章制度</summary>
            <div class="sub-body">
              <ol>
                <li>
                  <strong>《普通高等学校学生管理规定》</strong>：明确学生管理基本原则、学生权利与义务；涵盖学籍管理（入学注册、考核成绩、转专业转学、休学复学、退学、毕业结业）、校园秩序与课外活动、奖励与处分、学生申诉等全流程规范，是学生管理的根本遵循。
                </li>
                <li>
                  <strong>《高等学校学生行为准则》</strong>：从理想信念、爱国奉献、勤奋学习、遵纪守法、诚实守信、明礼修身、勤俭节约、身心健康八个方面划定学生行为规范。
                </li>
                <li>
                  <strong>《学生伤害事故处理办法》</strong>：规定学生伤害事故的责任认定、处理程序、损害赔偿及责任者处理，明确学校、学生、监护人及第三方责任边界。
                </li>
              </ol>
            </div>
          </details>

          <details class="sub-section">
            <summary>（二）学校章程</summary>
            <div class="sub-body">
              <ol>
                <li>
                  <strong>学校基本信息</strong>：中文简称莞工，英文缩写 DGUT；法定住所为东莞市松山湖科技产业园区大学路 1 号，设松山湖和莞城两个校区；校庆日为每年 4 月最后一个星期六。
                </li>
                <li>
                  <strong>办学定位与自主权</strong>：实行省市共建、以市为主的管理体制；具有制定规划、调节招生比例、设置专业、组织教学、开展科研与合作等办学自主权。
                </li>
                <li>
                  <strong>内部治理结构</strong>：实行党委领导下的校长负责制；明确党委、校长、纪委、学术委员会、学位评定委员会等机构职责；保障教职工代表大会、学生代表大会等民主管理形式。
                </li>
                <li>
                  <strong>师生权利与义务</strong>：明确教职工和学生在校期间合法权利与应履行义务，建立权益保护与救济机制。
                </li>
              </ol>
            </div>
          </details>

          <details class="sub-section">
            <summary>（三）教学管理制度</summary>
            <div class="sub-body">
              <ol>
                <li>
                  <strong>学籍管理</strong>：规范入学注册、考勤纪律、转专业与转学、休学复学留级、退学、学制与学习年限、毕业结业、学位授予、学业证书管理及成绩应用等；四年制学生最长学习年限为 6 年，需修满规定学分方可毕业。
                </li>
                <li>
                  <strong>学士学位授予</strong>：明确授予条件（遵纪守法、完成学业、平均学分绩点达标等）、不授予情形及授予程序；2025 年 1 月 1 日起依据《中华人民共和国学位法》执行。
                </li>
                <li>
                  <strong>课程考核管理</strong>：规定考核方式（考试、考查）、命题要求、试卷印制与保密、考试安排、免听缓考、评阅与成绩管理、考场纪律与监考职责，严肃考风考纪。
                </li>
                <li>
                  <strong>转专业管理</strong>：明确申请条件、不予转专业情形及程序；休学创业或退役复学学生可优先考虑转专业。
                </li>
                <li>
                  <strong>辅修制管理</strong>：设辅修学士学位与辅修专业两类；学生在校期间仅可修读一个辅修专业或辅修学士学位。
                </li>
                <li>
                  <strong>课外实践学分管理</strong>：全日制本科生需修满 15 个课外学分（含知行课程 2 学分、知行活动 4 学分）；本科插班生修满 5 个；学分涵盖德育教育、创新创业、特色体育、高雅美育、劳动教育、知行类等。
                </li>
                <li>
                  <strong>学生社区知行学院课外实践学分管理</strong>：知行学分共 6 个（课程 2、活动 4）；参军退伍复学可按规定认定学分。
                </li>
              </ol>
            </div>
          </details>

          <details class="sub-section">
            <summary>（四）学生管理制度</summary>
            <div class="sub-body">
              <ol>
                <li>
                  <strong>违纪处分办法</strong>：处分分为警告、严重警告、记过、留校察看、开除学籍五类；明确违纪行为处分标准、处分程序、学生权益及解除处分条件与程序。
                </li>
                <li>
                  <strong>学术不端行为处理</strong>：界定学术不端范围；建立教育预防、调查受理、认定处理及申诉复核机制，对不同情节制定处理措施并追责。
                </li>
                <li>
                  <strong>学生伤害事故处理</strong>：细化学校、学生及第三方责任，明确处理程序、损害赔偿及责任者处理方式。
                </li>
                <li>
                  <strong>国家奖助学金实施</strong>：国家奖学金（10000 元/年）、国家励志奖学金（6000 元/年）、国家助学金（2500-5000 元/年）；规定申请条件、评审程序与监督要求。
                </li>
                <li>
                  <strong>勤工助学管理</strong>：明确组织管理、岗位设置、活动管理与酬金标准，优先保障家庭经济困难学生参与。
                </li>
                <li>
                  <strong>学生社区管理</strong>：规范宿舍与安全管理，实行宿舍星级评比；对晚归、夜不归宿、使用违章电器等制定相应处理措施。
                </li>
                <li>
                  <strong>学生申诉管理</strong>：建立学生申诉处理委员会，明确申诉范围、程序、听证制度及复查决定送达方式。
                </li>
                <li>
                  <strong>大学生参军入伍管理</strong>：明确征兵组织领导、学业安排（保留学籍、免修部分课程、优先转专业等）、入伍激励与考核奖励。
                </li>
              </ol>
            </div>
          </details>

          <details class="sub-section">
            <summary>（五）其他管理制度</summary>
            <div class="sub-body">
              <ol>
                <li>
                  <strong>消防安全管理</strong>：贯彻预防为主，防消结合，明确组织与责任、检查整改、教育培训、应急预案与演练、经费及奖惩等。
                </li>
                <li>
                  <strong>校园交通管理</strong>：规范车辆与行人行为、交通设施与道路管理、事故处置与违规处理，保障校园交通秩序与安全。
                </li>
                <li>
                  <strong>电动自行车管理</strong>：实行上牌与封闭式出入管理；明确行驶、停放与充电规定及违规处理；禁止校外电动自行车驶入校园。
                </li>
                <li>
                  <strong>网络与信息安全管理</strong>：明确职责与安全防护要求，建立责任追究机制，严禁危害网络与信息安全的行为。
                </li>
                <li>
                  <strong>计算机网络管理</strong>：明确校园网范围与用户规范、运行维护、接入与 VPN、IP 与域名管理、安全管理及罚则。
                </li>
              </ol>
            </div>
          </details>
        </div>
      </details>

      <details class="section">
        <summary id="sec-3">三、附录内容</summary>
        <div class="section-body">
          <ol>
            <li>
              <strong>通识课程修读说明</strong>：明确通识课程（核心素养课、公共选修课）的培养目标、修读要求、学分标准及选课流程，不同年级学分要求有所差异。
            </li>
            <li>
              <strong>学生宿舍可使用电器种类</strong>：限定电脑、台灯、充电器等可用电器及规格；禁止电热毯、电磁炉等违章电器，强调购买正规产品并实名登记。
            </li>
            <li>
              <strong>校园网用户守则</strong>：规定校园网用户权利与义务，明确网络接入、使用与安全保密要求，对违规行为制定相应处理措施。
            </li>
          </ol>
        </div>
      </details>

      <details class="section">
        <summary id="sec-4">四、确认声明</summary>
        <div class="section-body">
          <p>学生需阅读并认可《东莞理工学院学生手册》内容，明确自身权利与义务后，由本人签名确认。</p>
        </div>
      </details>
    </div>
  </div>
</template>

<script setup>
const toc = [
  { id: "sec-1", title: "一、学校核心理念与定位" },
  { id: "sec-2", title: "二、五大核心制度板块" },
  { id: "sec-3", title: "三、附录内容" },
  { id: "sec-4", title: "四、确认声明" }
]

const scrollTo = (id) => {
  const el = document.getElementById(id)
  if (!el) return
  el.scrollIntoView({ behavior: "smooth", block: "start" })
}
</script>

<style scoped>
.student-handbook {
  max-width: 980px;
  margin: 0 auto;
  padding-bottom: 48px;
}

.handbook-hero {
  text-align: center;
  margin: 8px 0 14px 0;
}

.handbook-hero h2 {
  margin: 0 0 6px 0;
  color: #2c3e50;
  font-size: 1.65rem;
  font-weight: 800;
}

.handbook-hero p {
  margin: 0;
  color: #4f5d6b;
  font-size: 0.95rem;
}

.toc-card {
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(39, 174, 96, 0.18);
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 8px 28px rgba(39, 174, 96, 0.08);
  margin-bottom: 14px;
}

.toc-title {
  font-weight: 800;
  color: #2c3e50;
  margin-bottom: 10px;
}

.toc-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.toc-item {
  border: 1px solid rgba(39, 174, 96, 0.18);
  background: rgba(46, 204, 113, 0.08);
  color: #1f6f3c;
  border-radius: 999px;
  padding: 8px 12px;
  cursor: pointer;
  font-weight: 700;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.toc-item:hover {
  background: rgba(46, 204, 113, 0.14);
  border-color: rgba(39, 174, 96, 0.32);
}

.sections {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section,
.sub-section {
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(39, 174, 96, 0.18);
  border-radius: 16px;
  box-shadow: 0 8px 28px rgba(39, 174, 96, 0.08);
  overflow: hidden;
}

.sub-section {
  border-radius: 12px;
  box-shadow: none;
}

.section > summary,
.sub-section > summary {
  list-style: none;
  cursor: pointer;
  padding: 14px 16px;
  font-weight: 800;
  color: #2c3e50;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  scroll-margin-top: 90px;
}

.sub-section > summary {
  padding: 12px 14px;
  font-weight: 700;
  background: rgba(46, 204, 113, 0.06);
}

.section > summary::-webkit-details-marker,
.sub-section > summary::-webkit-details-marker {
  display: none;
}

.section > summary::after,
.sub-section > summary::after {
  content: "";
  width: 10px;
  height: 10px;
  border-right: 2px solid rgba(44, 62, 80, 0.7);
  border-bottom: 2px solid rgba(44, 62, 80, 0.7);
  transform: rotate(45deg);
  transition: transform 0.2s ease;
  flex-shrink: 0;
}

.section[open] > summary::after,
.sub-section[open] > summary::after {
  transform: rotate(-135deg);
}

.section-body {
  padding: 0 16px 16px 16px;
  color: #34495e;
  line-height: 1.7;
}

.section-body h4 {
  margin: 14px 0 8px 0;
  color: #1e5c35;
  font-size: 1rem;
}

.section-body p {
  margin: 10px 0;
}

.section-body ul,
.section-body ol {
  margin: 8px 0 8px 20px;
  padding: 0;
}

.section-body li {
  margin: 8px 0;
}

.sub-body {
  padding: 10px 14px 14px 14px;
  color: #34495e;
  line-height: 1.7;
}

@media (max-width: 640px) {
  .toc-item {
    width: 100%;
  }
}
</style>