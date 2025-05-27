// MongoDB初始化脚本
// 在Docker容器启动时自动执行

print('🚀 初始化Volcengine AI数据库...');

// 切换到目标数据库
db = db.getSiblingDB('volcengine_db');

// 创建用户集合并设置索引
print('📋 创建用户集合索引...');
db.users.createIndex(
    { "email": 1 }, 
    { 
        unique: true,
        name: "email_unique_index",
        background: true
    }
);

// 创建图像任务集合并设置索引
print('🎨 创建图像任务集合索引...');

// 用户ID索引（用于查询用户的任务）
db.image_tasks.createIndex(
    { "user_id": 1 }, 
    { 
        name: "user_id_index",
        background: true
    }
);

// 状态索引（用于查询特定状态的任务）
db.image_tasks.createIndex(
    { "status": 1 }, 
    { 
        name: "status_index",
        background: true
    }
);

// 创建时间索引（用于按时间排序）
db.image_tasks.createIndex(
    { "created": -1 }, 
    { 
        name: "created_desc_index",
        background: true
    }
);

// 复合索引：用户ID + 创建时间（用于用户任务列表查询）
db.image_tasks.createIndex(
    { "user_id": 1, "created": -1 }, 
    { 
        name: "user_created_index",
        background: true
    }
);

// 复合索引：状态 + 创建时间（用于任务队列管理）
db.image_tasks.createIndex(
    { "status": 1, "created": -1 }, 
    { 
        name: "status_created_index",
        background: true
    }
);

// 创建管理员用户（可选）
print('👤 创建管理员用户...');
try {
    db.users.insertOne({
        email: "admin@volcengine.ai",
        name: "系统管理员",
        role: "admin",
        created_at: new Date(),
        updated_at: new Date()
    });
    print('✅ 管理员用户创建成功');
} catch (e) {
    if (e.code === 11000) {
        print('ℹ️  管理员用户已存在');
    } else {
        print('❌ 创建管理员用户失败:', e.message);
    }
}

// 显示创建的索引
print('📊 用户集合索引:');
db.users.getIndexes().forEach(index => {
    print(`  - ${index.name}: ${JSON.stringify(index.key)}`);
});

print('📊 图像任务集合索引:');
db.image_tasks.getIndexes().forEach(index => {
    print(`  - ${index.name}: ${JSON.stringify(index.key)}`);
});

print('🎉 数据库初始化完成！'); 