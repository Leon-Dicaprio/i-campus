<template>
  <div class="student-handbook-view">
    <StudentHandbook />
  </div>
</template>

<script setup>
import StudentHandbook from '@/components/StudentHandbook.vue'
</script>

<style scoped>
.student-handbook-view {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #e8fff2 0%, #c9f7d6 45%, #b5e8c2 100%);
  padding: 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
</style>