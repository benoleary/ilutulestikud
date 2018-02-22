import { Component, Input } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { Observable } from 'rxjs/Rx';
import { Subscription } from 'rxjs/Subscription';
import { MatListModule } from '@angular/material/list';
import { IlutulestikudService } from '../ilutulestikud.service';
import { ChatMessage } from '../models/chatmessage.model'


@Component({
    selector: 'game-display',
    templateUrl: './gamedisplay.component.html',
  })
  export class GameDisplayComponent implements OnInit, OnDestroy
  {
    @Input() gameName: string;
    @Input() playerName: string;
    informationText: string;
    gameDataSubscription: Subscription;
    isAwaitingGameData: boolean;
    chatLog: ChatMessage[];

    constructor(public ilutulestikudService: IlutulestikudService)
    {
        this.gameName = null;
        this.playerName = null;
        this.chatLog = [];
    }

    ngOnInit(): void
    {
        this.gameDataSubscription =
          Observable
            .timer(0, 1000)
            .takeWhile(() => (this.gameName != null))
            .subscribe(
              () => this.refreshGameData(),
              thrownError => this.handleError(thrownError),
              () => {});
    }
  
    ngOnDestroy(): void
    {
      if (this.gameDataSubscription)
      {
        this.gameDataSubscription.unsubscribe();
      }
  
      this.gameName = null;
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
          .gameAsSeenByPlayer(this.gameName, this.playerName)
          .subscribe(
            fetchedGameData => this.displayGameData(fetchedGameData),
            thrownError => this.handleError(thrownError),
            () => {});
      }
    }

    displayGameData(fetchedGameData: Object): void
    {
        // If we have received game data to display, we are no longer waiting for the HTTP request to complete.
        this.isAwaitingGameData = false;
    
        // The object fetchedGameData["ChatLog"] is only an "array-like object",
        // not an array, so does not have foreach or length.
        const fetchedChatLog: Object[] = Array.from(fetchedGameData["ChatLog"]);

        // First of all we reduce the number of chat log lines if there were more than request gave us.
        if (this.chatLog.length > fetchedChatLog.length)
        {
            this.chatLog.length = fetchedChatLog.length;
        }
        
        for (var messageIndex: number = 0; messageIndex < fetchedChatLog.length; ++messageIndex)
        {
            const fetchedMessage: Object = fetchedChatLog[messageIndex];


            // We could replace each message with each refresh, but to avoid possible issues (such
            // as happens with the turn summaries), we update existing messages and only add new ones
            // when necessary.
            if (messageIndex < this.chatLog.length)
            {
                this.chatLog[messageIndex].refreshFromSource(fetchedMessage)
            }
            else
            {
                this.chatLog.push(new ChatMessage(fetchedMessage));
            }
        }
    }

    handleError(thrownError: Error): void
    {
      console.log("Error! " + JSON.stringify(thrownError));
      this.informationText = "Error! " + JSON.stringify(thrownError["error"]);
    }
  }