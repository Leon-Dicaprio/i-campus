<template>
  <div class="school-map-view">
  </div>
</template>

<script setup>
</script>

<style scoped>
.school-map-view {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #f2f4f7 0%, #e9edf2 100%);
  padding: 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.map-header {
  text-align: center;
}

.map-header h2 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 2rem;
  font-weight: 700;
}

.map-header p {
  margin: 0;
  color: #7f8c8d;
  font-size: 1.1rem;
}

.map-container {
  background: white;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

.map-placeholder {
  text-align: center;
  color: #7f8c8d;
}

.map-placeholder svg {
  margin-bottom: 16px;
}

.map-placeholder h3 {
  margin: 0 0 8px 0;
  font-size: 1.5rem;
  color: #2c3e50;
}

.map-placeholder p {
  margin: 0;
  font-size: 1rem;
}

.quick-locations h3 {
  margin: 0 0 20px 0;
  color: #2c3e50;
  font-size: 1.5rem;
  font-weight: 600;
}

.location-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.location-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  cursor: pointer;
}

.location-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.location-icon {
  font-size: 2.5rem;
  margin-bottom: 12px;
}

.location-card h4 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 1.1rem;
  font-weight: 600;
}

.location-card p {
  margin: 0;
  color: #7f8c8d;
  font-size: 0.9rem;
}

@media (max-width: 768px) {
  .school-map-view {
    padding: 16px;
  }
  
  .map-container {
    padding: 24px;
    min-height: 200px;
  }
  
  .location-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
  }
  
  .location-card {
    padding: 16px;
  }
  
  .location-icon {
    font-size: 2rem;
  }
}
</style>