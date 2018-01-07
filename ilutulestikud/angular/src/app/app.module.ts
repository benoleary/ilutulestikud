import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { InputTextModule } from 'primeng/primeng';

import { IlutulestikudModule } from '../ilutulestikud/ilutulestikud.module';
import { IlutulestikudComponent } from '../ilutulestikud/ilutulestikud.component';

import { AppComponent } from './app.component';


@NgModule({
  declarations: [
    AppComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    IlutulestikudModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
