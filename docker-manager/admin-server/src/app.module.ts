// admin-server/src/app.module.ts
@Module({
    imports: [
      TypeOrmModule.forRoot({
        type: 'sqlite',
        database: 'appstore.db',
        entities: [AppTemplate],
        synchronize: true,
      }),
    ],
    controllers: [AppTemplateController],
  })
  export class AppModule {}