## 架构和需求

- 创建一个Dcloud包
- 自定义Server类型，创建一个其类型的对象，将其挂载到 lsp.Server上
- 自定义Server的作用：
  1. 管理Dcloud相关操作的自定义数据
  2. 提供一些工具方法，比如获取或创建Dcloud自定义的LanguageService
  3. 以此为基点，覆写默认LanguageService的方法，就可以拦截默认操作，或添加自定义数据，为自定义功能提供支撑
- 现有的一些需求
  - 前提
    1. 默认的LaguageService从ts版本迁移到go版本后，LanguageService的使用方法变成了即用即创建，所有的数据都由Program进行管理
    2. 和ts版本一致的是，一个Program还是无法对包含的文件进行分包管理
    3. 所以想要实现自定义功能，就需要基于Program创建扩展出多个不同的Program进行文件库的管理，用的时候，再使用不同的Program创建不同的LanguageService进行功能Api调用
  - 需求
    1. 应该基于现有project创建自定义的Project，放到Server中管理起来（缓存），用时直接获取
    2. lsp.Server在handle某个Api时可以先从Server中通过project或者ls或者文件路径获取对应的LanguageService，获取到了就说明有我自定义的能力，就调用自定义的LanguageService的对应Api,否则还是调用默认的
    3. 对于同项目，不同场景，自定义LanguageService需要能创建多个ls进行分发
    4. 能自定义LanguageService也需要能自定义Program,对Program进行扩展，添加自定义数据，为自定义功能提供支撑
    5. 也需要能自定义ModuleResove,对ModuleResove进行扩展
    6. go多线程操作非常方便，需要想个办法，对多数据操作时，能进行并发处理，提升性能
    7. 其它


## 备忘录

- go中没有继续和代理，如果有接口的话还好，可能通过鸭子类型进行自定义类型传参,但是默认的LaguageService是类型，不是接口，lsp.Server中使用LanguageService的地方都是类型指针
- 所以如果要实现自定义功能改动原始代码最小的位置就是在lsp.Server去handle对应的Api时，尝试做拦截


## 下一步

-[x] 实现内存释放 参考`ProjectCollectionBuilder`中的释放逻辑