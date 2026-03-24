<template>
  <div class="campus-navigation">
    <div class="navigation-header">
      <h3>校园导航服务</h3>
      <p>快速找到您需要的位置</p>
    </div>
    
    <div class="search-section">
      <div class="search-box">
        <input 
          v-model="searchQuery" 
          @input="filterLocations"
          placeholder="搜索地点、建筑、设施..."
          class="search-input"
        />
        <svg class="search-icon" width="20" height="20" viewBox="0 0 24 24" fill="none">
          <path d="M21 21L16.5 16.5M19 11C19 15.4183 15.4183 19 11 19C6.58172 19 3 15.4183 3 11C3 6.58172 6.58172 3 11 3C15.4183 3 19 6.58172 19 11Z" 
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
    </div>

    <div class="location-categories">
      <div 
        v-for="category in categories" 
        :key="category.id"
        class="category-card"
        @click="selectCategory(category)"
        :class="{ active: selectedCategory === category.id }"
      >
        <div class="category-icon">
          <component :is="category.icon" />
        </div>
        <div class="category-info">
          <h4>{{ category.name }}</h4>
          <p>{{ category.description }}</p>
          <span class="location-count">{{ category.locations.length }} 个地点</span>
        </div>
      </div>
    </div>

    <div v-if="filteredLocations.length > 0" class="locations-list">
      <h4>{{ getCategoryName(selectedCategory) }}地点</h4>
      <div 
        v-for="location in filteredLocations" 
        :key="location.id"
        class="location-item"
        @click="selectLocation(location)"
      >
        <div class="location-info">
          <h5>{{ location.name }}</h5>
          <p>{{ location.description }}</p>
          <div class="location-details">
            <span class="floor" v-if="location.floor">{{ location.floor }}</span>
            <span class="hours" v-if="location.hours">{{ location.hours }}</span>
            <span class="distance" v-if="location.distance">{{ location.distance }}</span>
          </div>
        </div>
        <div class="location-actions">
          <button @click.stop="showDirections(location)" class="direction-btn">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
              <path d="M3 12L3 19L10 19M3 19L21 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            导航
          </button>
        </div>
      </div>
    </div>

    <div v-if="selectedLocation" class="location-detail">
      <div class="detail-header">
        <h4>{{ selectedLocation.name }}</h4>
        <button @click="closeDetail" class="close-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M18 6L6 18M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
      </div>
      <div class="detail-content">
        <p>{{ selectedLocation.description }}</p>
        <div class="detail-info">
          <div class="info-item" v-if="selectedLocation.floor">
            <strong>楼层：</strong>{{ selectedLocation.floor }}
          </div>
          <div class="info-item" v-if="selectedLocation.hours">
            <strong>开放时间：</strong>{{ selectedLocation.hours }}
          </div>
          <div class="info-item" v-if="selectedLocation.contact">
            <strong>联系方式：</strong>{{ selectedLocation.contact }}
          </div>
          <div class="info-item" v-if="selectedLocation.services">
            <strong>服务内容：</strong>{{ selectedLocation.services }}
          </div>
        </div>
        <div class="detail-actions">
          <button @click="showDirections(selectedLocation)" class="primary-btn">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
              <path d="M3 12L3 19L10 19M3 19L21 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            开始导航
          </button>
          <button @click="shareLocation(selectedLocation)" class="secondary-btn">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
              <path d="M4 12V8C4 6.89543 4.89543 6 6 6H8M16 6H18C19.1046 6 20 6.89543 20 8V12M20 20V16C20 14.8954 19.1046 14 18 14H16M8 14H6C4.89543 14 4 14.8954 4 16V20" 
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            分享位置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const searchQuery = ref('')
const selectedCategory = ref('all')
const selectedLocation = ref(null)

const categories = ref([
  {
    id: 'academic',
    name: '教学区域',
    description: '教学楼、实验室、教室',
    icon: 'AcademicIcon',
    locations: [
      { id: 1, name: '主教学楼A栋', description: '主要教学区域，包含多媒体教室', floor: '1-6层', hours: '7:00-22:00', distance: '步行2分钟' },
      { id: 2, name: '实验楼B栋', description: '物理、化学实验室', floor: '2-5层', hours: '8:00-18:00', distance: '步行5分钟' },
      { id: 3, name: '图书馆', description: '图书阅览、自习室', floor: '1-4层', hours: '8:00-22:00', distance: '步行3分钟' }
    ]
  },
  {
    id: 'dining',
    name: '餐饮服务',
    description: '食堂、餐厅、咖啡厅',
    icon: 'DiningIcon',
    locations: [
      { id: 4, name: '第一食堂', description: '中式快餐、特色窗口', floor: '1-2层', hours: '6:30-21:00', distance: '步行1分钟' },
      { id: 5, name: '第二食堂', description: '风味餐厅、清真食堂', floor: '1层', hours: '11:00-14:00, 17:00-20:00', distance: '步行4分钟' },
      { id: 6, name: '咖啡厅', description: '咖啡、轻食、休闲区', floor: '1层', hours: '8:00-22:00', distance: '步行2分钟' }
    ]
  },
  {
    id: 'dormitory',
    name: '宿舍区域',
    description: '学生宿舍、生活设施',
    icon: 'DormitoryIcon',
    locations: [
      { id: 7, name: '学生宿舍1号楼', description: '男生宿舍', floor: '1-6层', hours: '全天开放', distance: '步行8分钟' },
      { id: 8, name: '学生宿舍2号楼', description: '女生宿舍', floor: '1-6层', hours: '全天开放', distance: '步行10分钟' }
    ]
  },
  {
    id: 'service',
    name: '服务设施',
    description: '行政楼、医务室、商店',
    icon: 'ServiceIcon',
    locations: [
      { id: 9, name: '行政办公楼', description: '学生事务、教务处', floor: '1-4层', hours: '8:00-17:00', distance: '步行6分钟' },
      { id: 10, name: '校医院', description: '医疗服务、急救', floor: '1-2层', hours: '24小时急诊', distance: '步行7分钟' },
      { id: 11, name: '校园超市', description: '日用品、零食饮料', floor: '1层', hours: '8:00-22:00', distance: '步行3分钟' }
    ]
  },
  {
    id: 'sports',
    name: '运动设施',
    description: '体育馆、操场、健身房',
    icon: 'SportsIcon',
    locations: [
      { id: 12, name: '体育馆', description: '篮球、羽毛球、乒乓球', floor: '1-3层', hours: '9:00-21:00', distance: '步行5分钟' },
      { id: 13, name: '操场', description: '足球场、跑道', floor: '地面', hours: '6:00-21:00', distance: '步行4分钟' },
      { id: 14, name: '健身房', description: '健身器材、瑜伽室', floor: '2层', hours: '14:00-22:00', distance: '步行6分钟' }
    ]
  }
])

const filteredLocations = computed(() => {
  let locations = []
  
  if (selectedCategory.value === 'all') {
    locations = categories.value.flatMap(cat => cat.locations)
  } else {
    const category = categories.value.find(cat => cat.id === selectedCategory.value)
    locations = category ? category.locations : []
  }
  
  if (searchQuery.value) {
    locations = locations.filter(location => 
      location.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      location.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }
  
  return locations
})

const selectCategory = (category) => {
  selectedCategory.value = category.id
  selectedLocation.value = null
}

const selectLocation = (location) => {
  selectedLocation.value = location
}

const closeDetail = () => {
  selectedLocation.value = null
}

const showDirections = (location) => {
  // 这里可以集成实际的导航功能
  alert(`正在为您导航到：${location.name}`)
}

const shareLocation = (location) => {
  // 分享位置功能
  const shareText = `${location.name} - ${location.description}`
  if (navigator.share) {
    navigator.share({
      title: location.name,
      text: shareText
    })
  } else {
    // 复制到剪贴板
    navigator.clipboard.writeText(shareText)
    alert('位置信息已复制到剪贴板')
  }
}

const getCategoryName = (categoryId) => {
  const category = categories.value.find(cat => cat.id === categoryId)
  return category ? category.name : ''
}

const filterLocations = () => {
  // 搜索逻辑已在 computed 中处理
}

// 简单的图标组件
const AcademicIcon = () => h('svg', {width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none'}, [
  h('path', {d: 'M12 2L2 7L12 12L22 7L12 2Z', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M2 17L12 22L22 17', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M2 12L12 17L22 12', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'})
])

const DiningIcon = () => h('svg', {width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none'}, [
  h('path', {d: 'M3 2V7C3 8.10457 3.89543 9 5 9H19C20.1046 9 21 8.10457 21 7V2', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M3 7V18C3 19.1046 3.89543 20 5 20H19C20.1046 20 21 19.1046 21 18V7', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M8 22H16', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'})
])

const DormitoryIcon = () => h('svg', {width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none'}, [
  h('path', {d: 'M3 9L12 2L21 9V20C21 20.5304 20.7893 21.0391 20.4142 21.4142C20.0391 21.7893 19.5304 22 19 22H5C4.46957 22 3.96086 21.7893 3.58579 21.4142C3.21071 21.0391 3 20.5304 3 20V9Z', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M9 22V12H15V22', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'})
])

const ServiceIcon = () => h('svg', {width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none'}, [
  h('path', {d: 'M19 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H16L21 8V19C21 19.5304 20.7893 20.0391 20.4142 20.4142C20.0391 20.7893 19.5304 21 19 21Z', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M17 21V13H7V21', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M7 3V8H15', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'})
])

const SportsIcon = () => h('svg', {width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none'}, [
  h('circle', {cx: 12, cy: 12, r: 10, stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'}),
  h('path', {d: 'M12 6V12L16 14', stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round', 'stroke-linejoin': 'round'})
])
</script>

<style scoped>
.campus-navigation {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.navigation-header {
  text-align: center;
  margin-bottom: 30px;
}

.navigation-header h3 {
  font-size: 1.5rem;
  color: #2c3e50;
  margin-bottom: 8px;
}

.navigation-header p {
  color: #7f8c8d;
  font-size: 0.9rem;
}

.search-section {
  margin-bottom: 30px;
}

.search-box {
  position: relative;
  max-width: 500px;
  margin: 0 auto;
}

.search-input {
  width: 100%;
  padding: 12px 45px 12px 15px;
  border: 2px solid #e0e6ed;
  border-radius: 25px;
  font-size: 1rem;
  outline: none;
  transition: border-color 0.3s;
}

.search-input:focus {
  border-color: #3498db;
}

.search-icon {
  position: absolute;
  right: 15px;
  top: 50%;
  transform: translateY(-50%);
  color: #95a5a6;
}

.location-categories {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 15px;
  margin-bottom: 30px;
}

.category-card {
  background: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
}

.category-card:hover {
  border-color: #3498db;
  box-shadow: 0 4px 12px rgba(52, 152, 219, 0.1);
}

.category-card.active {
  border-color: #3498db;
  background: #f8f9fa;
}

.category-icon {
  width: 40px;
  height: 40px;
  background: #3498db;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  color: white;
}

.category-info h4 {
  margin: 0 0 5px 0;
  color: #2c3e50;
  font-size: 1.1rem;
}

.category-info p {
  margin: 0 0 8px 0;
  color: #7f8c8d;
  font-size: 0.85rem;
}

.location-count {
  color: #3498db;
  font-size: 0.8rem;
  font-weight: 500;
}

.locations-list h4 {
  margin-bottom: 15px;
  color: #2c3e50;
}

.location-item {
  background: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 10px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.location-item:hover {
  border-color: #3498db;
  box-shadow: 0 2px 8px rgba(52, 152, 219, 0.1);
}

.location-info h5 {
  margin: 0 0 5px 0;
  color: #2c3e50;
  font-size: 1rem;
}

.location-info p {
  margin: 0 0 8px 0;
  color: #7f8c8d;
  font-size: 0.85rem;
}

.location-details {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.location-details span {
  background: #f8f9fa;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  color: #6c757d;
}

.direction-btn {
  background: #3498db;
  color: white;
  border: none;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: background 0.3s;
}

.direction-btn:hover {
  background: #2980b9;
}

.location-detail {
  background: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 12px;
  padding: 20px;
  margin-top: 20px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.detail-header h4 {
  margin: 0;
  color: #2c3e50;
  font-size: 1.2rem;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 5px;
  color: #7f8c8d;
}

.detail-content p {
  color: #5a6c7d;
  margin-bottom: 15px;
  line-height: 1.5;
}

.detail-info {
  margin-bottom: 20px;
}

.info-item {
  margin-bottom: 8px;
  color: #5a6c7d;
}

.info-item strong {
  color: #2c3e50;
}

.detail-actions {
  display: flex;
  gap: 10px;
}

.primary-btn, .secondary-btn {
  padding: 10px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.3s;
  border: none;
}

.primary-btn {
  background: #3498db;
  color: white;
}

.primary-btn:hover {
  background: #2980b9;
}

.secondary-btn {
  background: #f8f9fa;
  color: #6c757d;
  border: 1px solid #e0e6ed;
}

.secondary-btn:hover {
  background: #e9ecef;
}
</style>
