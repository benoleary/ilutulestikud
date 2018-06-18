import { Component, Input, Output, EventEmitter } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { nullSafeIsEquivalent } from '@angular/compiler/src/output/output_ast';
import { Observable } from 'rxjs/Rx';
import { Subscription } from 'rxjs/Subscription';
import { MatListModule } from '@angular/material/list';
import { MatInputModule } from '@angular/material';
import { IlutulestikudService } from '../ilutulestikud.service';
import { LogMessage } from '../models/logmessage.model'
import { BackendIdentification } from '../models/backendidentification.model';



@Component({
    selector: 'game-display',
    templateUrl: './gamedisplay.component.html',
  })
  export class GameDisplayComponent implements OnInit, OnDestroy
  {
    @Input() gameIdentification: BackendIdentification;
    @Input() playerIdentification: BackendIdentification;
    @Output() onError: EventEmitter<any> = new EventEmitter();
    informationText: string;
    gameDataSubscription: Subscription;
    isAwaitingGameData: boolean;
    chatLog: LogMessage[];
    chatInput: string;

    constructor(public ilutulestikudService: IlutulestikudService)
    {
        this.gameIdentification = null;
        this.playerIdentification = null;
        this.chatLog = [];
        this.chatInput = null;
    }

    ngOnInit(): void
    {
        this.gameDataSubscription =
          Observable
            .timer(0, 1000)
            .takeWhile(() => (this.gameIdentification != null))
            .subscribe(
              () => this.refreshGameData(),
              thrownError => this.onError.emit(thrownError),
              () => {});
    }
  
    ngOnDestroy(): void
    {
      if (this.gameDataSubscription)
      {
        this.gameDataSubscription.unsubscribe();
      }
  
      this.gameIdentification = null;
    }

    refreshGameData(): void
    {
      // We only request new game data if we are not waiting for the response to the last request.
      if (!this.isAwaitingGameData)
      {
        // We note that we are now awaiting the HTTP response (this.isAwaitingGameData will be set
        // back to false by this.displayGameData(fetchedGameData) which will run when we get the
        // response to the request).
        this.isAwaitingGameData = true;
        this.ilutulestikudService
          .gameAsSeenByPlayer(this.gameIdentification, this.playerIdentification)
          .subscribe(
            fetchedGameData => this.parseGameData(fetchedGameData),
            thrownError => this.onError.emit(thrownError),
            () => {});
      }
    }

    parseGameData(fetchedGameData: Object): void
    {
        // If we have received game data to display, we are no longer waiting for the HTTP request to complete.
        this.isAwaitingGameData = false;
    
        // The object fetchedGameData["ChatLog"] is only an "array-like object",
        // not an array, so does not have foreach or length.
        LogMessage.refreshListFromSource(this.chatLog, Array.from(fetchedGameData["ChatLog"]))
    }

    sendChat(): void
    {
        if (this.chatInput)
        {
            this.ilutulestikudService
            .sendChatMessage(this.gameIdentification, this.playerIdentification, this.chatInput)
            .subscribe(
                () => {},
                thrownError => this.onError.emit(thrownError),
                () => {});
            this.chatInput = null;
        }
    }
  }